# procwatch

A small Linux command-line tool, written in Go, that reports system health by reading `/proc`.
No dependencies beyond the Go standard library.

```
$ procwatch snapshot
uptime   73.2 hours
load     0.46 0.11 0.04
memory   4.9% used (392 of 8022 MB)
disk /   88.5% used (28 GB free)
```

## Commands

| Command | What it does |
|---|---|
| `procwatch snapshot [--json] [--disk PATH]` | Uptime, load average, memory and disk usage |
| `procwatch top [--json] [-n COUNT]` | Processes using the most resident memory |
| `procwatch ports [--json]` | Listening TCP sockets (IPv4 and IPv6) |
| `procwatch check [--max-mem PCT] [--max-disk PCT] [--max-load N]` | Threshold check; exits 2 if one fails |
| `procwatch serve [--addr :8080]` | Web status page and JSON API (`/api/snapshot`, `/api/top`, `/api/ports`, `/healthz`) |

## How it works
- Reads `/proc/meminfo`, `/proc/loadavg`, `/proc/uptime`, `/proc/<pid>/status` and `/proc/net/tcp{,6}`.
- Every parser takes plain text, so the tests run on fixed sample input without a real system.
- IPv4 addresses in `/proc/net/tcp` are little-endian hex; they are decoded to dotted form.
- Processes that exit while being read are skipped instead of causing an error.
- Disk usage comes from `statfs`.

## Shell scripts
`scripts/health.sh` runs a snapshot and a check, and prints the top memory users when a check fails.
It uses `set -euo pipefail` and passes the exit code through so cron or CI can act on it.
Limits can be set with `MAX_MEM` and `MAX_DISK`.

## Build, test, run
```
make build
make test
make health
```
Requires Go 1.22 or newer. Linux only.

## Web mode and deployment
`procwatch serve` shows the same data in a small web page and as JSON. It reports the machine it runs on, so a hosted copy shows the host's stats, not yours.
- **Render:** `Dockerfile` and `render.yaml` are included. Create a Web Service from this repo with the Docker runtime. The free plan sleeps when idle, so the first request can be slow.
- **Vercel:** `api/index.go` is a Go serverless function that serves the same handler, with `vercel.json` rewriting all paths to it. Values come from the serverless container, so they reflect that environment.

## Packaging
`snap/snapcraft.yaml` defines a snap with the `system-observe` and `network-observe` plugs.
It has not been built or published yet; build it with `snapcraft`.

## Layout
```
cmd/procwatch/        command-line entry point (snapshot, top, ports, check, serve)
web/                  HTTP handler, JSON API and status page (with tests)
api/index.go          Vercel serverless entry point
Dockerfile, render.yaml  Render deployment
internal/sysinfo/     /proc parsers and collectors, with tests
scripts/health.sh     Bash wrapper for cron/CI
snap/snapcraft.yaml   snap packaging definition
.github/workflows/    CI: gofmt, vet, tests, shellcheck, build
```

## License
MIT
