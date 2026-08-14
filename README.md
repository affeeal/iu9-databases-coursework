# Dgraph Dataset Converter and Benchmark Coursework

A reproducible Dgraph coursework project built around a configurable Go
CSV/TSV-to-RDF converter, four public graph datasets, DQL schemas and workloads,
a local Docker Compose environment, and a small query-measurement utility. The
historical evaluation was completed in 2024; this repository exposes the code
and results without presenting them as a current database benchmark.

## Highlights

- YAML-driven declarations, casts, RDF rules, and edge facets;
- explicit header/configuration validation and atomically written output;
- deterministic conversion of multiple datasets concurrently;
- Dgraph RDF output with escaped string literals;
- loopback-only local Dgraph services pinned to v23.1.1;
- separate Dgraph latency, client wall time, and a clearly labeled host-memory
  proxy.

## Architecture

```text
CSV/TSV + convert.yml
          |
          v
   Go converter ----> output.rdf + schema.dql
                              |
                              v
                       Dgraph v23.1.1
                              |
                              v
                    DQL query + benchmark
```

## Quick start

Prerequisites are Go 1.22+, Docker Engine, Docker Compose v2, `curl`, and GNU
Make. The Go-only path does not require Docker or any large dataset.

### Converter-only demo

The repository includes a two-row fixture with facets, quoted text, a backslash,
and a multiline field:

```sh
go run ./cmd/converter \
  -dataset-path internal/converter/testdata/minimal
diff -u \
  internal/converter/testdata/minimal/expected.rdf \
  internal/converter/testdata/minimal/output.rdf
go test -race ./...
```

The beginning of the generated RDF is:

```rdf
_:p1 <name> "Alice" (note="hello \"graph\"\\world", score=7) .
_:p1 <note> "hello \"graph\"\\world" .
_:p1 <score> "7" .
_:p1 <external_id> "1" .
```

`make demo` wraps the conversion and comparison.

### End-to-end Dgraph demo

```sh
docker compose up --detach
scripts/wait-for-dgraph.sh

make demo
scripts/load-dataset.sh internal/converter/testdata/minimal
go run ./cmd/benchmark -- \
  -query-path internal/converter/testdata/minimal/query.dql \
  -print-response

docker compose down
```

`docker compose down` preserves the named data volume. The repository has no
ordinary command that deletes it. Published Dgraph ports bind to `127.0.0.1`,
and the Compose network has a fixed gateway so the admin whitelist can be
limited to localhost and that gateway.

`DGRAPH_VERSION=v23.1.1` in [.env](.env) pins the historical database version.
Ratel is intentionally absent from the default environment rather than using a
floating UI image.

## Datasets and workloads

Original source data are not redistributed. Upstream providers control their
licensing and access terms; inspect those terms before downloading or reusing
the data.

| Dataset | Source format and graph interpretation | Repository artifacts |
|---|---|---|
| [MOOC User Action Dataset](https://snap.stanford.edu/data/act-mooc.html) | TSV; users and course targets connected through actions | [configuration](datasets/act-mooc/convert.yml), [schema](datasets/act-mooc/schema.dql), [queries](datasets/act-mooc/queries) |
| [Elliptic++ Transactions Dataset](https://github.com/git-disl/EllipticPlusPlus/tree/main/Transactions%20Dataset) | CSV; Bitcoin transactions and transaction-flow edges | [configuration](datasets/elliptic++/convert.yml), [schema](datasets/elliptic++/schema.dql), [queries](datasets/elliptic++/queries) |
| [California road network](https://snap.stanford.edu/data/roadNet-CA.html) | Tab-separated edge list; road intersections and directed links | [configuration](datasets/roadNet-CA/convert.yml), [schema](datasets/roadNet-CA/schema.dql), [queries](datasets/roadNet-CA/queries) |
| [Stablecoin ERC20 Transactions Dataset](https://snap.stanford.edu/data/ERC20-stablecoins.html) | CSV; addresses, transfer nodes, and sender/recipient/contract edges | [configuration](datasets/ERC20-stablecoins/convert.yml), [schema](datasets/ERC20-stablecoins/schema.dql), [queries](datasets/ERC20-stablecoins/queries) |

See [datasets/README.md](datasets/README.md) for the expected local layout.

## Historical coursework results

The report records five runs per query on an Intel Core i5-10300H machine with
16 GiB RAM and a local cluster containing one Zero and one Alpha node. The table
below reproduces the reported Dgraph request latency in milliseconds without
altering the values:

| Dataset | Query 1 | Query 2 | Query 3 | Query 4 | Query 5 |
|---|---:|---:|---:|---:|---:|
| Elliptic++ Transactions | 27 | 85 | 448 | 3,072 | 23 |
| MOOC User Actions | 8,030 | 202 | 1,235 | 2,989 | 2,667 |
| California road network | 8,849 | 1 | 1 | 13,661 | 30,540 |
| Stablecoin ERC20 Transactions | 4,199 | 1,093 | 1,099 | 56,025 | 5,546 |

The queries differ across heterogeneous datasets, so rows are not
apples-to-apples comparisons. These are historical coursework measurements,
not current Dgraph results. The report's memory figure is the fall in free RAM
for the entire host during a query; it is sensitive to unrelated processes and
sampling frequency and is not Dgraph process RSS.

See the [PDF report](paper/paper.pdf) and the [measurement source](
paper/4_estimation.tex) for the full Russian-language methodology, topology,
disk-space, latency, and memory tables.

## Converter configuration

The complete reference is in [docs/converter-config.md](docs/converter-config.md).
A minimal rule looks like this:

```yaml
files:
  - name: people.csv
    declarations:
      - name: person
        type: id
        prefix: p
      - name: name
        type: string
    rdfs:
      - subject: person
        predicate: name
        object: name
```

Run one dataset or every dataset directory beneath a root:

```sh
go run ./cmd/converter -dataset-path path/to/dataset
go run ./cmd/converter -datasets-path datasets
```

## Benchmark CLI

```sh
go run ./cmd/benchmark -- \
  -query-path datasets/roadNet-CA/queries/query2.dql \
  -host localhost \
  -port 9080 \
  -sample-interval 100ms \
  -query-timeout 30s \
  -print-response
```

See [cmd/benchmark/README.md](cmd/benchmark/README.md) for field definitions,
units, timeout behavior, and measurement caveats.

## Repository layout

```text
cmd/converter/       converter CLI
cmd/benchmark/       DQL query and measurement CLI
internal/converter/  YAML validation and conversion pipeline
internal/rdf/        RDF terms, escaping, facets, and triples
datasets/            public-dataset configs, DQL schemas, and workloads
docs/                converter configuration reference
scripts/             readiness and Live Loader helpers
paper/               original Russian coursework report
compose.yml          pinned local Zero + Alpha environment
```

## Testing

```sh
make check
```

This runs a `gofmt` check, `go vet ./...`, `go test -race ./...`, and
`docker compose config --quiet`. Unit tests cover sampler shutdown and timeout,
malformed responses, configuration validation, Unicode delimiters, duplicate
and missing headers, RDF escaping, stable facets, atomic output, and the full
minimal conversion.

## Limitations and provenance

- The four workloads were designed for different graph structures; the project
  is not a general database benchmarking framework.
- The system-wide free-RAM proxy is useful only as a rough observation.
- The local environment is deliberately pinned to historical Dgraph v23.1.1.
- Large source datasets and generated RDF are ignored and must remain local.
- The Go source and report were written for the coursework. Dgraph, Go modules,
  Docker, the report template/emblem, and upstream datasets retain their own
  licenses and provenance.
- No repository-wide license has been added pending an explicit ownership and
  licensing decision for all retained material.
