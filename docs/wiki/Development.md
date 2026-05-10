# Development

## Project Structure

- `cmd/mcserver`: executable entrypoint.
- `internal/cli`: command definitions.
- `internal/core`: Minecraft server core providers.
- `internal/modrinth`: Modrinth API client.
- `internal/download`: built-in downloader.
- `internal/mirror`: GitHub proxy rewrite support.

## Checks

```bash
go test ./...
go build -o mcserver ./cmd/mcserver
```

## Release

Push a tag such as `v0.1.0` to run the release workflow:

```bash
git tag v0.1.0
git push origin v0.1.0
```

The workflow builds cross-platform binaries and attaches them to a GitHub Release.
