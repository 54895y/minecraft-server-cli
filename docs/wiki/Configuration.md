# Configuration

Print the config path:

```bash
mcserver config path
```

List active settings:

```bash
mcserver config list
```

Set defaults:

```bash
mcserver config set download.threads 16
mcserver config set core.source official
mcserver config set github.proxy none
```

Reset the config file:

```bash
mcserver config reset
```

Important keys:

- `download.threads`
- `download.timeout`
- `download.retries`
- `download.user_agent`
- `core.source`
- `github.proxy`
- `paths.output_dir`
