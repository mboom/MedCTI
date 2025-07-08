package main

import (
	"context"
	"flag"
	"fmt"
	"time"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"github.com/golang/protobuf/ptypes/empty"
	blockchain "github.com/mboom/MedCTI/blockchain/proto"
)

var (
	host = flag.String("host", "localhost", "The hostname or IP address that will be used to listen.")
	port = flag.Int("port", 50051, "The server port")
)

// create connection to the blockchain simulator
func connect() (*grpc.ClientConn, blockchain.BlockchainClient) {
	conn, err := grpc.NewClient(fmt.Sprintf("%v:%d", *host, *port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	ledger := blockchain.NewBlockchainClient(conn)
	return conn, ledger
}

func read(ctx context.Context, ledger blockchain.BlockchainClient) {
	// create blockchain stream reader
	data, err := ledger.FetchLogData(ctx, &empty.Empty{})
	if err != nil {
		panic(err)
	}

	// receive data from blockchain
	for err == nil {
		flow, err := data.Recv()
		if err != nil {
			break
		}

		// print data
		fmt.Println(flow)
	}
}

func main() {
	// parse command line arguments
	flag.Parse()

	// create connection
	conn, ledger := connect()
	defer conn.Close()

	// create context
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// read blockchain
	read(ctx, ledger)
}
