# Download Sources

## Official

Paper, Folia, and Velocity use PaperMC Downloads Service v3.

Purpur uses the PurpurMC v2 API.

## Mirrors

The CLI includes mirror source names for:

- `fastmirror`
- `msl`

Mirror providers first resolve metadata through the official provider, then rewrite the download URL through the mirror layer. If a mirror endpoint changes, use:

```bash
mcserver core download --source official ...
```

## GitHub Proxy

For arbitrary GitHub download URLs:

```bash
mcserver get https://github.com/owner/repo/releases/download/v1/file.zip --github-proxy gh-proxy
```

Supported rewrite names:

- `none`
- `gh-proxy`
- `akams`
- `gitproxy`
- `gitwarp`
- `custom`

For `custom`, provide:

```bash
--github-proxy-url https://proxy.example/
```
