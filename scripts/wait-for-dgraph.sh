#!/usr/bin/env bash

set -euo pipefail

timeout_seconds=${DGRAPH_WAIT_TIMEOUT:-60}
if [[ ! $timeout_seconds =~ ^[1-9][0-9]{0,5}$ ]]; then
  echo "DGRAPH_WAIT_TIMEOUT must be a positive integer of at most six digits" >&2
  exit 2
fi
deadline=$((SECONDS + timeout_seconds))
health_url=${DGRAPH_HEALTH_URL:-http://127.0.0.1:8080/health}

while true; do
  remaining=$((deadline - SECONDS))
  if ((remaining <= 0)); then
    echo "Dgraph did not become ready within ${timeout_seconds}s" >&2
    exit 1
  fi
  if curl --fail --silent --show-error --max-time "$remaining" "$health_url" >/dev/null 2>&1; then
    break
  fi
  if ((SECONDS < deadline)); then
    sleep 1
  fi
done

echo "Dgraph Alpha is ready at $health_url"
