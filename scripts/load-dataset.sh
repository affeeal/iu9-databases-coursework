#!/usr/bin/env bash

set -euo pipefail

if (($# != 1)); then
  echo "Usage: $0 <dataset-directory>" >&2
  exit 2
fi

repo_root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
dataset_path=$(realpath -- "$1")
for required_file in convert.yml output.rdf schema.dql; do
  if [[ ! -f $dataset_path/$required_file ]]; then
    echo "Missing dataset file: $dataset_path/$required_file" >&2
    exit 1
  fi
done

cd -- "$repo_root"
docker compose run --rm --no-deps \
  --volume "$dataset_path:/dataset:ro" \
  alpha dgraph live \
  --files /dataset/output.rdf \
  --schema /dataset/schema.dql \
  --alpha alpha:9080 \
  --zero zero:5080
