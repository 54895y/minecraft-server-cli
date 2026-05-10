package download

import "testing"

func TestSplitRanges(t *testing.T) {
	ranges := splitRanges(10, 3)
	want := []byteRange{{0, 2}, {3, 5}, {6, 9}}
	if len(ranges) != len(want) {
		t.Fatalf("got %d ranges", len(ranges))
	}
	for i := range want {
		if ranges[i] != want[i] {
			t.Fatalf("range %d got %+v want %+v", i, ranges[i], want[i])
		}
	}
}

func TestParseContentRange(t *testing.T) {
	if got := parseContentRange("bytes 0-0/12345"); got != 12345 {
		t.Fatalf("got %d", got)
	}
}
