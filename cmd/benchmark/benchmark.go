package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"github.com/dgraph-io/dgo/v230"
	"github.com/dgraph-io/dgo/v230/protos/api"
	"github.com/pbnjay/memory"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type cancelFunc func()

const (
	defaultDuration     = time.Duration(time.Millisecond * 100)
	defaultHost         = "localhost"
	defaultPort         = 9080
	defaultPrintRespond = false
	grpcMaxRecieveBytes = 1e+9
)

var (
	duration     time.Duration
	host         string
	port         uint
	printRespond bool
	queryPath    string
)

func formTarget(host string, port uint) string {
	return fmt.Sprintf("%s:%d", host, port)
}

func prettyPrintJson(src []byte) {
	var prettyJson bytes.Buffer
	err := json.Indent(&prettyJson, src, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(prettyJson.Bytes()))
}

func getDgraphClient(host string, port uint) (*dgo.Dgraph, cancelFunc) {
	conn, err := grpc.Dial(
		formTarget(host, port),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(grpcMaxRecieveBytes),
		),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal(err)
	}

	dc := api.NewDgraphClient(conn)
	return dgo.NewDgraphClient(dc), func() {
		if err := conn.Close(); err != nil {
			log.Fatal(err)
		}
	}
}

func performQuery(dg *dgo.Dgraph, query string, duration time.Duration, printRespond bool) {
	memoryBefore := memory.FreeMemory()
	var memoryMinimum uint64 = math.MaxUint64
	quit := make(chan bool)
	go func() {
		for {
			select {
			case <-quit:
				break
			default:
				freeMemory := memory.FreeMemory()
				if freeMemory < memoryMinimum {
					memoryMinimum = freeMemory
				}
				time.Sleep(duration)
			}
		}
	}()

	txn := dg.NewReadOnlyTxn().BestEffort()
	resp, err := txn.Query(context.Background(), query)
	if err != nil {
		log.Fatal(err)
	}
	quit <- true

	log.Printf("Free RAM before the execution: %d bytes.\n", memoryBefore)
	log.Printf("Free RAM minimum during the execution: %d bytes.\n", memoryMinimum)
	log.Printf("Free RAM consumption: %d bytes.\n", memoryBefore-memoryMinimum)
	log.Printf("Request latency: %d nanoseconds.\n", resp.Latency.GetTotalNs())

	if printRespond {
		prettyPrintJson(resp.GetJson())
	}
}

func validateInput() string {
	if queryPath == "" {
		log.Fatal("Please, specify DQL query file path")
	}

	file, err := os.ReadFile(queryPath)
	if err != nil {
		log.Fatal(err)
	}

	return string(file)
}

func init() {
	flag.DurationVar(&duration, "duration", defaultDuration, "Time duration between free memory measurements")
	flag.StringVar(&host, "host", defaultHost, "Dgraph server host")
	flag.UintVar(&port, "port", defaultPort, "Dgraph server port")
	flag.BoolVar(&printRespond, "print-respond", defaultPrintRespond, "Print JSON query respond")
	flag.StringVar(&queryPath, "query-path", "", "DQL query file path")
}

func main() {
	flag.Parse()

	query := validateInput()
	dg, cancel := getDgraphClient(host, port)
	defer cancel()

	performQuery(dg, query, duration, printRespond)
}
