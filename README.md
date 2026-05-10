# Minecraft Server CLI

`minecraft server cli` is a Go command-line tool for downloading Minecraft server cores, Modrinth projects, and arbitrary files with a built-in multi-threaded downloader.

This public project was created with assistance from ChatGPT.

## Features

- Download Paper, Folia, Velocity, and Purpur server cores.
- Resolve PaperMC downloads through the official Downloads Service v3.
- Search and download Modrinth mods/plugins by loader and Minecraft version.
- Download any direct URL with configurable multi-threaded HTTP range downloads.
- Built-in GitHub download proxy URL rewriting for common acceleration services.
- Persistent config through `mcserver config`.
- GitHub Actions builds Windows, Linux, and macOS binaries.

## Install

Download a release artifact from GitHub Actions or Releases, then place `mcserver` on your PATH.

For local builds:

```bash
go build -o mcserver ./cmd/mcserver
```

## Quick Start

```bash
mcserver core list --type paper
mcserver core builds --type paper --mc 1.21.10
mcserver core download --type paper --mc 1.21.10 --build latest -o server.jar
```

```bash
mcserver modrinth search viaversion --loader paper --type mod
mcserver modrinth download viaversion --loader paper --search -o plugins
```

```bash
mcserver get https://example.com/file.jar -o downloads/file.jar --threads 16
```

## Core Sources

Official source:

```bash
mcserver core download --type paper --source official --mc latest -o server.jar
```

`latest` prefers the newest traditional Minecraft version such as `1.21.x`. If PaperMC exposes a non-traditional version like `26.1.2`, pass it explicitly with `--mc 26.1.2`.

Mirror source:

```bash
mcserver core download --type paper --source fastmirror --mc 1.21.10 -o server.jar
mcserver core download --type paper --source msl --mc 1.21.10 -o server.jar
```

FastMirror and MSL support is implemented through a mirror URL layer. If a mirror changes its public API, use `--source official` or update the mirror template in code/config.

## GitHub Proxy Downloads

```bash
mcserver get https://github.com/owner/repo/releases/download/v1/file.zip --github-proxy gh-proxy
mcserver get https://github.com/owner/repo/releases/download/v1/file.zip --github-proxy akams
mcserver get https://github.com/owner/repo/releases/download/v1/file.zip --github-proxy custom --github-proxy-url https://your-proxy.example/
```

Built-in proxy names:

- `none`
- `gh-proxy`
- `akams`
- `gitproxy`
- `gitwarp`

`moretools` is documented as a user-facing proxy directory, but it is not hardcoded as a direct download rewrite endpoint.

## Configuration

```bash
mcserver config path
mcserver config list
mcserver config set download.threads 16
mcserver config set core.source official
mcserver config reset
```

Default config path:

- Windows: `%APPDATA%\mcserver\config.yaml`
- Linux/macOS: the OS user config directory under `mcserver/config.yaml`

## API Notes

PaperMC requires a clear non-generic `User-Agent` for Downloads Service requests. The default user agent identifies this project and can be changed:

```bash
mcserver config set download.user_agent "your-name/minecraft-server-cli/1.0 (https://github.com/your-name/repo)"
```

Modrinth also requires a uniquely identifying `User-Agent`.

## Development

```bash
go test ./...
go build -o mcserver ./cmd/mcserver
```

The main code is split into:

- `internal/cli`: Cobra command wiring.
- `internal/core`: server core providers.
- `internal/modrinth`: Modrinth API client and file selection.
- `internal/download`: multi-thread downloader.
- `internal/mirror`: GitHub proxy URL rewriting.

## License

MIT
