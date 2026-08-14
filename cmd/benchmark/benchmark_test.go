package main

import (
	"bytes"
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
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

func TestReadQueryReportsMissingFile(t *testing.T) {
	if _, err := readQuery(t.TempDir() + "/missing.dql"); err == nil {
		t.Fatal("missing query unexpectedly succeeded")
	}
}
