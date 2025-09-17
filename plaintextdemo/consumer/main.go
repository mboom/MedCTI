package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/google/uuid"
	blockchain "github.com/mboom/MedCTI/blockchain/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	bc_host      = flag.String("host", "localhost", "The hostname or IP address that will be used to connect to a blockchain simulator.")
	bc_port      = flag.Int("port", 50051, "The TCP port number of the blockchain simulator.")
	csp_host     = flag.String("host", "localhost", "The hostname or IP address that will be used to connect to a cryptographic service porvider.")
	csp_port     = flag.Int("port", 50052, "The TCP port number of the CSP.")
	localAddress = flag.Int("localAddress", 0, "The 4-byte host address in the simulated network environment")
)

// create connection to the blockchain simulator
func connect() (*grpc.ClientConn, blockchain.BlockchainClient) {
	conn, err := grpc.NewClient(fmt.Sprintf("%v:%d", *bc_host, *bc_port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	ledger := blockchain.NewBlockchainClient(conn)
	return conn, ledger
}

// publish a flow
func publishFlow(ctx context.Context, ledger blockchain.BlockchainClient, flow *blockchain.Flow) error {
	_, err := ledger.PublishLogData(ctx, flow)
	if err != nil {
		return err
	}

	return nil
}

// generate randomly a network flow
func generateFlow(kid uint32) *blockchain.Flow {
	// choose a random remote address and make sure it is really a remote address
	remote := rand.IntN(16)
	for remote == *localAddress {
		remote = rand.IntN(16)
	}

	// choose a random direction
	direction, source, destination := rand.IntN(2), []byte{byte(*localAddress)}, []byte{byte(*localAddress)}

	// set the remote address as source or destination based on the chosen direction
	switch direction {
	case 0:
		source = []byte{byte(remote)}
	default:
		destination = []byte{byte(remote)}
	}

	// return a new flow
	return &blockchain.Flow{Id: uuid.New().ID(), Kid: kid, Destination: destination, Source: source}
}

func main() {
	// parse command line arguments
	flag.Parse()

	// create connection
	conn, ledger := connect()
	defer conn.Close()

	// create context
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// prepare kid
	counter, kid := 0, uuid.New().ID()

	// generate traffic
	for {
		// wait a while
		time.Sleep(time.Duration(rand.IntN(3)) * time.Second)

		// new kid after every 1000 flows
		if counter >= 1000 {
			kid = uuid.New().ID()
		}
		counter++

		// create a random flow
		flow := generateFlow(kid)

		// publish a the flow to the ledger
		err := publishFlow(ctx, ledger, flow)

		if err != nil {
			break
		}
	}
}
