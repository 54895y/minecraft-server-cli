package download

import (
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/54895y/minecraft-server-cli/internal/httpx"
)

type Options struct {
	URL      string
	Output   string
	Threads  int
	Checksum Checksum
	Progress io.Writer
}

type Checksum struct {
	Algorithm string
	Value     string
}

type Result struct {
	Output string
	Bytes  int64
}

type Downloader struct {
	client *httpx.Client
}

func New(client *httpx.Client) *Downloader {
	return &Downloader{client: client}
}

func (d *Downloader) Download(ctx context.Context, opts Options) (Result, error) {
	if opts.URL == "" {
		return Result{}, fmt.Errorf("download URL is required")
	}
	if opts.Output == "" {
		return Result{}, fmt.Errorf("output path is required")
	}
	if opts.Threads <= 0 {
		opts.Threads = 1
	}
	if err := ensureOutputDir(opts.Output); err != nil {
		return Result{}, err
	}

	size, ranges, err := d.probe(ctx, opts.URL)
	if err != nil {
		return Result{}, err
	}
	if size <= 0 || !ranges || opts.Threads == 1 {
		return d.single(ctx, opts)
	}
	return d.parallel(ctx, opts, size)
}

func (d *Downloader) probe(ctx context.Context, rawURL string) (int64, bool, error) {
	req, err := d.client.NewRequest(ctx, http.MethodHead, rawURL, nil)
	if err != nil {
		return 0, false, err
	}
	resp, err := d.client.Do(req)
	if err != nil || resp.StatusCode == http.StatusMethodNotAllowed {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
		return d.probeWithRange(ctx, rawURL)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return 0, false, fmt.Errorf("probe failed: %s", resp.Status)
	}
	size := resp.ContentLength
	acceptRanges := strings.Contains(strings.ToLower(resp.Header.Get("Accept-Ranges")), "bytes")
	return size, acceptRanges && size > 0, nil
}

func (d *Downloader) probeWithRange(ctx context.Context, rawURL string) (int64, bool, error) {
	req, err := d.client.NewRequest(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return 0, false, err
	}
	req.Header.Set("Range", "bytes=0-0")
	resp, err := d.client.Do(req)
	if err != nil {
		return 0, false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusPartialContent {
		return parseContentRange(resp.Header.Get("Content-Range")), true, nil
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return resp.ContentLength, false, nil
	}
	return 0, false, fmt.Errorf("range probe failed: %s", resp.Status)
}

func (d *Downloader) single(ctx context.Context, opts Options) (Result, error) {
	req, err := d.client.NewRequest(ctx, http.MethodGet, opts.URL, nil)
	if err != nil {
		return Result{}, err
	}
	resp, err := d.client.Do(req)
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Result{}, fmt.Errorf("download failed: %s", resp.Status)
	}

	tmp := opts.Output + ".tmp"
	out, err := os.Create(tmp)
	if err != nil {
		return Result{}, err
	}
	defer out.Close()

	w := io.Writer(out)
	var written atomic.Int64
	if opts.Progress != nil {
		w = io.MultiWriter(out, progressWriter{written: &written})
		stop := reportProgress(ctx, opts.Progress, &written, resp.ContentLength)
		defer stop()
	}

	if _, err := io.Copy(w, resp.Body); err != nil {
		return Result{}, err
	}
	if err := out.Close(); err != nil {
		return Result{}, err
	}
	if err := verifyFile(tmp, opts.Checksum); err != nil {
		return Result{}, err
	}
	if err := os.Rename(tmp, opts.Output); err != nil {
		return Result{}, err
	}
	info, _ := os.Stat(opts.Output)
	if info == nil {
		return Result{Output: opts.Output}, nil
	}
	return Result{Output: opts.Output, Bytes: info.Size()}, nil
}

func (d *Downloader) parallel(ctx context.Context, opts Options, size int64) (Result, error) {
	ranges := splitRanges(size, opts.Threads)
	tmpDir := opts.Output + ".parts"
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return Result{}, err
	}
	defer os.RemoveAll(tmpDir)

	var written atomic.Int64
	stop := func() {}
	if opts.Progress != nil {
		stop = reportProgress(ctx, opts.Progress, &written, size)
	}
	defer stop()

	errCh := make(chan error, len(ranges))
	var wg sync.WaitGroup
	for i, r := range ranges {
		wg.Add(1)
		go func(idx int, rr byteRange) {
			defer wg.Done()
			if err := d.downloadPart(ctx, opts.URL, filepath.Join(tmpDir, fmt.Sprintf("%04d.part", idx)), rr, &written); err != nil {
				errCh <- err
			}
		}(i, r)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return Result{}, err
		}
	}

	tmp := opts.Output + ".tmp"
	out, err := os.Create(tmp)
	if err != nil {
		return Result{}, err
	}
	for i := range ranges {
		part, err := os.Open(filepath.Join(tmpDir, fmt.Sprintf("%04d.part", i)))
		if err != nil {
			_ = out.Close()
			return Result{}, err
		}
		if _, err := io.Copy(out, part); err != nil {
			_ = part.Close()
			_ = out.Close()
			return Result{}, err
		}
		_ = part.Close()
	}
	if err := out.Close(); err != nil {
		return Result{}, err
	}
	if err := verifyFile(tmp, opts.Checksum); err != nil {
		return Result{}, err
	}
	if err := os.Rename(tmp, opts.Output); err != nil {
		return Result{}, err
	}
	return Result{Output: opts.Output, Bytes: size}, nil
}

func (d *Downloader) downloadPart(ctx context.Context, rawURL, output string, r byteRange, written *atomic.Int64) error {
	req, err := d.client.NewRequest(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", r.Start, r.End))
	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("range %d-%d returned %s", r.Start, r.End, resp.Status)
	}
	out, err := os.Create(output)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, io.TeeReader(resp.Body, progressWriter{written: written}))
	return err
}

type byteRange struct {
	Start int64
	End   int64
}

func splitRanges(size int64, threads int) []byteRange {
	if threads <= 1 || size <= 0 {
		return []byteRange{{Start: 0, End: size - 1}}
	}
	if int64(threads) > size {
		threads = int(size)
	}
	chunk := size / int64(threads)
	ranges := make([]byteRange, 0, threads)
	var start int64
	for i := 0; i < threads; i++ {
		end := start + chunk - 1
		if i == threads-1 {
			end = size - 1
		}
		ranges = append(ranges, byteRange{Start: start, End: end})
		start = end + 1
	}
	return ranges
}

func parseContentRange(value string) int64 {
	idx := strings.LastIndex(value, "/")
	if idx < 0 || idx == len(value)-1 {
		return -1
	}
	n, err := strconv.ParseInt(value[idx+1:], 10, 64)
	if err != nil {
		return -1
	}
	return n
}

type progressWriter struct {
	written *atomic.Int64
}

func (p progressWriter) Write(b []byte) (int, error) {
	n := len(b)
	p.written.Add(int64(n))
	return n, nil
}

func reportProgress(ctx context.Context, out io.Writer, written *atomic.Int64, total int64) func() {
	done := make(chan struct{})
	go func() {
		t := time.NewTicker(500 * time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-done:
				printProgress(out, written.Load(), total)
				_, _ = fmt.Fprintln(out)
				return
			case <-ctx.Done():
				return
			case <-t.C:
				printProgress(out, written.Load(), total)
			}
		}
	}()
	return func() { close(done) }
}

func printProgress(out io.Writer, written, total int64) {
	if total > 0 {
		_, _ = fmt.Fprintf(out, "\rdownloaded %s / %s (%3.0f%%)", humanBytes(written), humanBytes(total), float64(written)*100/float64(total))
		return
	}
	_, _ = fmt.Fprintf(out, "\rdownloaded %s", humanBytes(written))
}

func humanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for x := n / unit; x >= unit; x /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(n)/float64(div), "KMGTPE"[exp])
}

func verifyFile(path string, checksum Checksum) error {
	if checksum.Value == "" {
		return nil
	}
	algorithm := strings.ToLower(strings.TrimSpace(checksum.Algorithm))
	var h hash.Hash
	switch algorithm {
	case "sha1":
		h = sha1.New()
	case "sha256":
		h = sha256.New()
	case "sha512":
		h = sha512.New()
	default:
		return fmt.Errorf("unsupported checksum algorithm %q", checksum.Algorithm)
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	got := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(got, checksum.Value) {
		return fmt.Errorf("%s mismatch: got %s want %s", algorithm, got, checksum.Value)
	}
	return nil
}

func absOrClean(path string) string {
	if dir := filepath.Dir(path); dir != "." {
		return path
	}
	return filepath.Join(".", path)
}

func ensureOutputDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}
