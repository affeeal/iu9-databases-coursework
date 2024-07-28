package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/dgraph-io/dgo/v230"
	"github.com/dgraph-io/dgo/v230/protos/api"
	"github.com/pbnjay/memory"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type cancelFunc func()

type lstVar []string

const (
	defaultHost         = "localhost"
	defaultPort         = 9080
	grpcMaxRecieveBytes = 1e+9
)

var (
	host           string
	port           uint
	queryFilenames lstVar
)

func (lst *lstVar) String() string {
	return fmt.Sprint(*lst)
}

func (lst *lstVar) Set(value string) error {
	if len(*lst) > 0 {
		return errors.New("queries flag is already set")
	}

	*lst = strings.Split(value, ",")
	return nil
}

func formTarget(host string, port uint) string {
	return fmt.Sprintf("%s:%d", host, port)
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
	dg := dgo.NewDgraphClient(dc)

	return dg, func() {
		if err := conn.Close(); err != nil {
			log.Fatal(err)
		}
	}
}

func performQuery(dg *dgo.Dgraph, queryFilename string) {
	log.Println("Handling", queryFilename)
	query, err := os.ReadFile(queryFilename)
	if err != nil {
		log.Fatal(err)
	}

	txn := dg.NewReadOnlyTxn().BestEffort()
	memoryBefore := memory.FreeMemory()
	resp, err := txn.Query(context.Background(), string(query))
	memoryAfter := memory.FreeMemory()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Free RAM before the execution: %d bytes\n", memoryBefore)
	log.Printf("Free RAM after the execution: %d bytes\n", memoryAfter)
	log.Printf("RAM change: %d bytes", memoryBefore-memoryAfter)

	latency := resp.Latency
	log.Printf("Request latency: %d nanoseconds\n", latency.GetTotalNs())
}

func init() {
	flag.Var(
		&queryFilenames,
		"queries",
		"comma-separated list of DQL query filenames",
	)
	flag.StringVar(&host, "host", defaultHost, "dgraph server host")
	flag.UintVar(&port, "port", defaultPort, "dgraph server port")
}

func main() {
	flag.Parse()

	dg, cancel := getDgraphClient(host, port)
	defer cancel()

	for _, queryFilename := range queryFilenames {
		performQuery(dg, queryFilename)
	}
}
