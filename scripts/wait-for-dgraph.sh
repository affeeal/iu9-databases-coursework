#!/usr/bin/env bash

set -euo pipefail

timeout_seconds=${DGRAPH_WAIT_TIMEOUT:-60}
deadline=$((SECONDS + timeout_seconds))
health_url=${DGRAPH_HEALTH_URL:-http://127.0.0.1:8080/health}

while ! curl --fail --silent --show-error "$health_url" >/dev/null 2>&1; do
  if ((SECONDS >= deadline)); then
    echo "Dgraph did not become ready within ${timeout_seconds}s" >&2
    exit 1
  fi
  sleep 1
done

echo "Dgraph Alpha is ready at $health_url"
