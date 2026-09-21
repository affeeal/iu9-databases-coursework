package main

import (
	"errors"
	"flag"
	"io"
	"testing"
)

func TestRunRejectsInvalidArguments(t *testing.T) {
	for _, args := range [][]string{
		nil,
		{"-dataset-path", "one", "-datasets-path", "all"},
		{"-dataset-path", "one", "extra"},
		{"-unknown"},
	} {
		if err := run(args, io.Discard); err == nil {
			t.Errorf("run(%q) unexpectedly succeeded", args)
		}
	}
}

func TestRunHelp(t *testing.T) {
	if err := run([]string{"-help"}, io.Discard); !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("expected help without opening a dataset, got %v", err)
	}
}
