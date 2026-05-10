# Commands

## Core

List Minecraft versions:

```bash
mcserver core list --type paper
```

List builds:

```bash
mcserver core builds --type paper --mc 1.21.10
```

Download a server core:

```bash
mcserver core download --type paper --mc 1.21.10 --build latest -o server.jar
```

Supported core types:

- `paper`
- `folia`
- `velocity`
- `purpur`

## Modrinth

Search:

```bash
mcserver modrinth search viaversion --loader paper
```

Download by slug or ID:

```bash
mcserver modrinth download viaversion --loader paper -o plugins
```

Download by search query:

```bash
mcserver modrinth download "Geyser" --search --loader paper --mc 1.21.10 -o plugins
```

## Any URL

```bash
mcserver get https://example.com/file.jar -o downloads/file.jar --threads 16
```
