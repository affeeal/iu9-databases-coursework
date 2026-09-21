package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"io"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dgraph-io/dgo/v230/protos/api"
)

func TestMeasureOperationRejectsNonPositiveInterval(t *testing.T) {
	for _, interval := range []time.Duration{0, -time.Nanosecond} {
		_, _, err := measureOperation(
			context.Background(), interval, func() uint64 { return 1 },
			func(context.Context) error { return nil },
		)
		if err == nil {
			t.Fatalf("measureOperation(%s) unexpectedly succeeded", interval)
		}
	}
}

func TestMeasureOperationHandlesImmediateCompletion(t *testing.T) {
	var samples atomic.Uint64
	measurement, _, err := measureOperation(
		context.Background(), time.Hour,
		func() uint64 {
			samples.Add(1)
			return 42
		},
		func(context.Context) error { return nil },
	)
	if err != nil {
		t.Fatal(err)
	}
	if measurement.before != 42 || measurement.minimum != 42 {
		t.Fatalf("unexpected measurement: %+v", measurement)
	}
	if samples.Load() != 1 {
		t.Fatalf("expected one real sample, got %d", samples.Load())
	}
}

func TestMeasureOperationStopsSampler(t *testing.T) {
	var samples atomic.Uint64
	_, _, err := measureOperation(
		context.Background(), time.Millisecond,
		func() uint64 { return 100 - samples.Add(1) },
		func(context.Context) error {
			time.Sleep(4 * time.Millisecond)
			return nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	afterReturn := samples.Load()
	time.Sleep(4 * time.Millisecond)
	if samples.Load() != afterReturn {
		t.Fatal("sampler continued after measureOperation returned")
	}
}

func TestMeasureOperationPropagatesTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()
	_, _, err := measureOperation(
		ctx, time.Millisecond, func() uint64 { return 1 },
		func(operationContext context.Context) error {
			<-operationContext.Done()
			return operationContext.Err()
		},
	)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline error, got %v", err)
	}
}

func TestPrettyPrintJSONRejectsMalformedResponse(t *testing.T) {
	if _, err := prettyPrintJSON([]byte("{")); err == nil {
		t.Fatal("malformed JSON unexpectedly succeeded")
	}
}

func TestParseOptionsValidation(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "missing path"},
		{name: "empty host", args: []string{"-query-path", "query.dql", "-host", ""}},
		{name: "zero port", args: []string{"-query-path", "query.dql", "-port", "0"}},
		{name: "large port", args: []string{"-query-path", "query.dql", "-port", "65536"}},
		{name: "zero interval", args: []string{"-query-path", "query.dql", "-sample-interval", "0s"}},
		{name: "zero timeout", args: []string{"-query-path", "query.dql", "-query-timeout", "0s"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := parseOptions(test.args, &bytes.Buffer{}); err == nil {
				t.Fatal("invalid options unexpectedly succeeded")
			}
		})
	}
}

func TestFormTarget(t *testing.T) {
	for _, test := range []struct{ host, want string }{
		{"localhost", "localhost:9080"},
		{"127.0.0.1", "127.0.0.1:9080"},
		{"::1", "[::1]:9080"},
	} {
		if got := formTarget(test.host, 9080); got != test.want {
			t.Errorf("formTarget(%q) = %q, want %q", test.host, got, test.want)
		}
	}
}

type failingWriter struct{ err error }

func (w failingWriter) Write([]byte) (int, error) { return 0, w.err }

func TestPrintResultPropagatesWriteError(t *testing.T) {
	want := errors.New("output closed")
	result := queryResult{response: &api.Response{}}
	if err := printResult(failingWriter{want}, result, false); !errors.Is(err, want) {
		t.Fatalf("expected write error, got %v", err)
	}
}

func TestPrintResult(t *testing.T) {
	result := queryResult{
		response:   &api.Response{Json: []byte(`{"nodes":[]}`), Latency: &api.Latency{TotalNs: 25}},
		memory:     memoryMeasurement{before: 100, minimum: 80},
		clientTime: 50 * time.Nanosecond,
	}
	var output bytes.Buffer
	if err := printResult(&output, result, true); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"System-wide free-RAM drop proxy: 20 bytes",
		"Dgraph-reported latency: 25 nanoseconds",
		"Client wall-clock duration: 50 nanoseconds",
		"  \"nodes\": []",
	} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("missing %q in output %q", want, output.String())
		}
	}
}

func TestPrintResultRejectsInvalidResponse(t *testing.T) {
	for _, response := range []*api.Response{nil, {Json: []byte("{")}} {
		var output bytes.Buffer
		if err := printResult(&output, queryResult{response: response}, true); err == nil {
			t.Fatal("invalid response unexpectedly succeeded")
		}
		if output.Len() != 0 {
			t.Fatalf("invalid response produced a partial report: %s", &output)
		}
	}
}

func TestRunHelp(t *testing.T) {
	if err := run([]string{"-help"}, io.Discard); !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("expected help without a query or connection, got %v", err)
	}
}

func TestReadQueryReportsMissingFile(t *testing.T) {
	if _, err := readQuery(t.TempDir() + "/missing.dql"); err == nil {
		t.Fatal("missing query unexpectedly succeeded")
	}
}
