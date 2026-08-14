# DQL Query Measurement Utility

The benchmark command performs one best-effort, read-only DQL query and prints
two timing measurements plus a rough system-wide memory proxy. It is intended
for reproducing the coursework workflow, not for profiler-quality process
accounting.

```sh
go run ./cmd/benchmark -- \
  -query-path datasets/roadNet-CA/queries/query2.dql \
  -host localhost \
  -port 9080 \
  -sample-interval 100ms \
  -query-timeout 30s \
  -print-response
```

## Flags

- `-query-path` (required): DQL query file;
- `-host` (default `localhost`): Dgraph gRPC host;
- `-port` (default `9080`): Dgraph gRPC port in the range 1 through 65535;
- `-sample-interval` (default `100ms`): positive interval between free-host-RAM
  samples;
- `-query-timeout` (default `30s`): positive upper bound for the query;
- `-print-response` (default `false`): pretty-print the JSON response.

The old misspelled `-print-respond` flag has been removed; invalid flags and
values return a nonzero status.

## Output

- `Host free RAM before query` (bytes): the initial host-wide sample;
- `Minimum host free RAM during query` (bytes): the lowest sampled value;
- `System-wide free-RAM drop proxy` (bytes): the difference between those two
  values;
- `Dgraph-reported latency` (nanoseconds): Dgraph's response latency field;
- `Client wall-clock duration` (nanoseconds): elapsed time around the client
  request, including client-side overhead.

The sampler takes an initial real sample, uses a ticker for subsequent samples,
and is always stopped and joined before results are read. A query that completes
before the first tick therefore reports a valid zero drop rather than an
underflow.

The RAM figure observes the entire host. Unrelated processes, filesystem cache,
and the chosen sampling interval all affect it; it is not isolated Dgraph RSS.
Use a quiet machine or constrained container environment when comparing runs.
Timeout, connection, query, transaction cleanup, malformed response, and output
errors are reported on stderr with a nonzero status.
