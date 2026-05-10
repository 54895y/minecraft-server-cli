# Installation

## GitHub Actions Artifacts

Every push to `main` runs the build workflow and uploads binaries for:

- Windows amd64
- Linux amd64
- macOS amd64
- macOS arm64

Download the artifact for your platform and place `mcserver` on your PATH.

## Build From Source

```bash
git clone https://github.com/54895y/minecraft-server-cli.git
cd minecraft-server-cli
go build -o mcserver ./cmd/mcserver
```

## Verify

```bash
mcserver version
mcserver --help
```
