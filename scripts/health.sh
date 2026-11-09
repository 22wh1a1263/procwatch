#!/usr/bin/env bash
# Run procwatch checks and report in a form cron or CI can use.
# Usage: scripts/health.sh [path-to-procwatch]
# Exit codes: 0 healthy, 1 error, 2 a check failed.
set -euo pipefail

bin="${1:-./bin/procwatch}"
max_mem="${MAX_MEM:-90}"
max_disk="${MAX_DISK:-90}"

if [[ ! -x "$bin" ]]; then
  echo "error: $bin not found or not executable (run 'make build')" >&2
  exit 1
fi

echo "== snapshot =="
"$bin" snapshot

echo "== check (mem<=${max_mem}%, disk<=${max_disk}%) =="
# 'check' exits 2 on failure; capture it instead of letting set -e abort.
status=0
"$bin" check --max-mem "$max_mem" --max-disk "$max_disk" || status=$?

if [[ $status -eq 2 ]]; then
  echo "unhealthy: top memory users:" >&2
  "$bin" top -n 5 >&2
fi
exit "$status"
