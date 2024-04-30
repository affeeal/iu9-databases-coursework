package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/dgraph-io/dgo/v230"
	"github.com/dgraph-io/dgo/v230/protos/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type cancelFunc func()

type lstVar []string

const (
	defaultHost = "localhost"
	defaultPort = 9080
	testQuery   = `
		{
			q(func: has(successors), first: 1000) {
				uid	
			}
		}
	`
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
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		log.Fatal(err)
	}

	dc := api.NewDgraphClient(conn)
	dg := dgo.NewDgraphClient(dc)

	// TODO: login

	return dg, func() {
		if err := conn.Close(); err != nil {
			log.Fatal(err)
		}
	}
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

	txn := dg.NewReadOnlyTxn().BestEffort()
	resp, err := txn.Query(
		context.Background(),
		testQuery,
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp.Latency.GetProcessingNs())
}
