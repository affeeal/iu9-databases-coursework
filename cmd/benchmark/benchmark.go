package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dgraph-io/dgo/v230"
	"github.com/dgraph-io/dgo/v230/protos/api"
	"github.com/pbnjay/memory"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	defaultSampleInterval = 100 * time.Millisecond
	defaultQueryTimeout   = 30 * time.Second
	defaultHost           = "localhost"
	defaultPort           = 9080
	grpcMaxReceiveBytes   = 1_000_000_000
)

type options struct {
	host           string
	port           int
	queryPath      string
	sampleInterval time.Duration
	queryTimeout   time.Duration
	printResponse  bool
}

type memoryMeasurement struct {
	before  uint64
	minimum uint64
}

func (measurement memoryMeasurement) dropProxy() uint64 {
	return measurement.before - measurement.minimum
}

type queryResult struct {
	response   *api.Response
	memory     memoryMeasurement
	clientTime time.Duration
}

func formTarget(host string, port int) string {
	return net.JoinHostPort(host, strconv.Itoa(port))
}

func prettyPrintJSON(source []byte) (string, error) {
	var formatted bytes.Buffer
	if err := json.Indent(&formatted, source, "", "  "); err != nil {
		return "", fmt.Errorf("format response JSON: %w", err)
	}
	return formatted.String(), nil
}

func parseOptions(arguments []string, output io.Writer) (options, error) {
	var result options
	flags := flag.NewFlagSet("benchmark", flag.ContinueOnError)
	flags.SetOutput(output)
	flags.DurationVar(
		&result.sampleInterval,
		"sample-interval",
		defaultSampleInterval,
		"interval between host free-RAM samples",
	)
	flags.DurationVar(
		&result.queryTimeout,
		"query-timeout",
		defaultQueryTimeout,
		"maximum Dgraph query duration",
	)
	flags.StringVar(&result.host, "host", defaultHost, "Dgraph server host")
	flags.IntVar(&result.port, "port", defaultPort, "Dgraph gRPC port")
	flags.BoolVar(
		&result.printResponse,
		"print-response",
		false,
		"print formatted JSON query response",
	)
	flags.StringVar(&result.queryPath, "query-path", "", "DQL query file path")

	if err := flags.Parse(arguments); err != nil {
		return options{}, err
	}
	if flags.NArg() != 0 {
		return options{}, fmt.Errorf("unexpected positional arguments: %v", flags.Args())
	}
	if result.queryPath == "" {
		return options{}, errors.New("query-path is required")
	}
	if strings.TrimSpace(result.host) == "" {
		return options{}, errors.New("host must not be empty")
	}
	if result.port < 1 || result.port > 65535 {
		return options{}, fmt.Errorf("port must be in the range 1..65535: %d", result.port)
	}
	if result.sampleInterval <= 0 {
		return options{}, errors.New("sample-interval must be positive")
	}
	if result.queryTimeout <= 0 {
		return options{}, errors.New("query-timeout must be positive")
	}
	return result, nil
}

func readQuery(path string) (string, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read query %q: %w", path, err)
	}
	return string(contents), nil
}

func getDgraphClient(host string, port int) (*dgo.Dgraph, *grpc.ClientConn, error) {
	connection, err := grpc.NewClient(
		formTarget(host, port),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(grpcMaxReceiveBytes),
		),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("create Dgraph connection: %w", err)
	}

	client := api.NewDgraphClient(connection)
	return dgo.NewDgraphClient(client), connection, nil
}

func measureOperation(
	ctx context.Context,
	interval time.Duration,
	sample func() uint64,
	operation func(context.Context) error,
) (memoryMeasurement, time.Duration, error) {
	if interval <= 0 {
		return memoryMeasurement{}, 0, errors.New("sample interval must be positive")
	}

	initial := sample()
	measurement := memoryMeasurement{before: initial, minimum: initial}
	samplerContext, stopSampler := context.WithCancel(ctx)
	samplerDone := make(chan struct{})
	go func() {
		defer close(samplerDone)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-samplerContext.Done():
				return
			case <-ticker.C:
				if freeMemory := sample(); freeMemory < measurement.minimum {
					measurement.minimum = freeMemory
				}
			}
		}
	}()

	started := time.Now()
	err := operation(ctx)
	elapsed := time.Since(started)
	stopSampler()
	<-samplerDone
	return measurement, elapsed, err
}

func query(
	ctx context.Context,
	dgraph *dgo.Dgraph,
	dql string,
) (response *api.Response, err error) {
	transaction := dgraph.NewReadOnlyTxn().BestEffort()
	defer func() {
		discardErr := transaction.Discard(context.Background())
		if discardErr != nil {
			err = errors.Join(err, fmt.Errorf("discard transaction: %w", discardErr))
		}
	}()
	return transaction.Query(ctx, dql)
}

func performQuery(
	ctx context.Context,
	dgraph *dgo.Dgraph,
	dql string,
	sampleInterval time.Duration,
) (queryResult, error) {
	var response *api.Response
	measurement, clientTime, err := measureOperation(
		ctx,
		sampleInterval,
		memory.FreeMemory,
		func(operationContext context.Context) error {
			var queryErr error
			response, queryErr = query(operationContext, dgraph, dql)
			return queryErr
		},
	)
	if err != nil {
		return queryResult{}, fmt.Errorf("execute query: %w", err)
	}
	return queryResult{
		response:   response,
		memory:     measurement,
		clientTime: clientTime,
	}, nil
}

func printResult(output io.Writer, result queryResult, printResponse bool) error {
	if result.response == nil {
		return errors.New("Dgraph returned no response")
	}
	var report bytes.Buffer
	fmt.Fprintf(&report, "Host free RAM before query: %d bytes\n", result.memory.before)
	fmt.Fprintf(&report, "Minimum host free RAM during query: %d bytes\n", result.memory.minimum)
	fmt.Fprintf(
		&report,
		"System-wide free-RAM drop proxy: %d bytes\n",
		result.memory.dropProxy(),
	)
	fmt.Fprintf(&report, "Dgraph-reported latency: %d nanoseconds\n", result.response.GetLatency().GetTotalNs())
	fmt.Fprintf(&report, "Client wall-clock duration: %d nanoseconds\n", result.clientTime.Nanoseconds())

	if printResponse {
		formatted, err := prettyPrintJSON(result.response.GetJson())
		if err != nil {
			return err
		}
		fmt.Fprintln(&report, formatted)
	}
	if _, err := report.WriteTo(output); err != nil {
		return fmt.Errorf("write query result: %w", err)
	}
	return nil
}

func run(arguments []string, output io.Writer) (err error) {
	configuration, err := parseOptions(arguments, output)
	if err != nil {
		return err
	}
	dql, err := readQuery(configuration.queryPath)
	if err != nil {
		return err
	}
	dgraph, connection, err := getDgraphClient(configuration.host, configuration.port)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := connection.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close Dgraph connection: %w", closeErr))
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), configuration.queryTimeout)
	defer cancel()
	result, err := performQuery(ctx, dgraph, dql, configuration.sampleInterval)
	if err != nil {
		return err
	}
	return printResult(output, result, configuration.printResponse)
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		log.Printf("benchmark: %v", err)
		os.Exit(1)
	}
}
