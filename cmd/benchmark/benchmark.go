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
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type cancelFunc func()

type lstVar []string

const (
	defaultHost         = "localhost"
	defaultPort         = 9080
	grpcMaxRecieveBytes = 1e+8
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

func performQuery(dg *dgo.Dgraph, queryFilename string) *api.Latency {
	query, err := os.ReadFile(queryFilename)
	if err != nil {
		log.Fatal(err)
	}

	txn := dg.NewReadOnlyTxn().BestEffort()
	resp, err := txn.Query(context.Background(), string(query))
	if err != nil {
		log.Fatal(err)
	}

	return resp.GetLatency()
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

	var latency *api.Latency
	for _, queryFilename := range queryFilenames {
		latency = performQuery(dg, queryFilename)
		fmt.Println(*latency)
		// TODO: handle latency
	}
}
