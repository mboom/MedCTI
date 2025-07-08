package main

import (
	"context"
	"flag"
	"fmt"
	"time"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"github.com/golang/protobuf/ptypes/empty"
	blockchain "github.com/mboom/MedCTI/blockchain/proto"
)

var (
	host = flag.String("host", "localhost", "The hostname or IP address that will be used to listen.")
	port = flag.Int("port", 50051, "The server port")
)

func main() {
	// parse command line arguments
	flag.Parse()

	// create connection to the blockchain simulator
	conn, err := grpc.NewClient(fmt.Sprintf("%v:%d", *host, *port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	ledger := blockchain.NewBlockchainClient(conn)

	// publish a message
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err = ledger.PublishLogData(ctx, TapTraffic())
	if err != nil {
		panic(err)
	}

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

func TapTraffic() *blockchain.Flow {
	flow := blockchain.Flow{Id: uuid.New().ID(), Kid: uuid.New().ID(), Destination: []byte{2}, Source: []byte{1} }
	return &flow
}
