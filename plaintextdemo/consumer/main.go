package main

import (
	"context"
	"flag"
	"fmt"
	"time"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

// publish a flow
func publish(ctx context.Context, ledger blockchain.BlockchainClient) {
	_, err := ledger.PublishLogData(ctx, &blockchain.Flow{Id: uuid.New().ID(), Kid: uuid.New().ID(), Destination: []byte{2}, Source: []byte{1}})
	if err != nil {
		panic(err)
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

	// publish a message
	publish(ctx, ledger)
}
