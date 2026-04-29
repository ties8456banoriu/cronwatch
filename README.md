# cronwatch

Lightweight daemon that monitors cron job execution times and sends alerts on missed or slow runs.

## Installation

```bash
go install github.com/yourname/cronwatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourname/cronwatch.git && cd cronwatch && go build ./...
```

## Usage

Define your monitored jobs in a `cronwatch.yaml` config file:

```yaml
jobs:
  - name: daily-backup
    schedule: "0 2 * * *"
    timeout: 30m
    alert:
      email: ops@example.com

  - name: hourly-sync
    schedule: "0 * * * *"
    timeout: 5m
    alert:
      webhook: https://hooks.example.com/alert
```

Start the daemon:

```bash
cronwatch --config cronwatch.yaml
```

Wrap your existing cron commands to report execution status:

```bash
# In your crontab
0 2 * * * cronwatch exec --job daily-backup -- /usr/local/bin/backup.sh
```

cronwatch will alert you if a job:
- Does not start within its expected schedule window
- Exceeds its configured timeout
- Exits with a non-zero status code

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--config` | `cronwatch.yaml` | Path to config file |
| `--log-level` | `info` | Log verbosity (`debug`, `info`, `warn`) |
| `--dry-run` | `false` | Validate config without starting daemon |

## License

MIT — see [LICENSE](LICENSE) for details.