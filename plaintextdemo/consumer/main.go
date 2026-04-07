package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"math/rand/v2"
	"os"
	"time"

	"github.com/google/uuid"
	blockchain "github.com/mboom/MedCTI/blockchain/proto"
	csp "github.com/mboom/MedCTI/csp/plaintextdemo/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	bc_host      = flag.String("bcHost", "localhost", "The hostname or IP address that will be used to connect to a blockchain simulator.")
	bc_port      = flag.Int("bcPort", 50051, "The TCP port number of the blockchain simulator.")
	csp_host     = flag.String("cspHost", "localhost", "The hostname or IP address that will be used to connect to a cryptographic service porvider.")
	csp_port     = flag.Int("cspPort", 50052, "The TCP port number of the CSP.")
	localAddress = flag.Int("localAddress", 0, "The 4-byte host address in the simulated network environment.")
	netflows     = flag.String("netflows", "../data/fs1000-LITNET-2020.csv", "The dataset that contains recorded netflows.")
	kidLog       = flag.String("kidLog", "../data/kid-consumer.csv", "The log file that will contain the key identifiers used to publish data.")
)

// create connection to the blockchain simulator
func connectLedger() (*grpc.ClientConn, blockchain.BlockchainClient) {
	conn, err := grpc.NewClient(fmt.Sprintf("%v:%d", *bc_host, *bc_port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	ledger := blockchain.NewBlockchainClient(conn)
	return conn, ledger
}

// create connection to the cryptographic service provider
func connectCsp() (*grpc.ClientConn, csp.CSPClient) {
	conn, err := grpc.NewClient(fmt.Sprintf("%v:%d", *csp_host, *csp_port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	csp := csp.NewCSPClient(conn)
	return conn, csp
}

// publish a flow
func publishFlow(ctx context.Context, ledger blockchain.BlockchainClient, flow *blockchain.Flow) error {
	_, err := ledger.PublishLogData(ctx, flow)
	if err != nil {
		return err
	}

	return nil
}

// load recorded network flows form file
func loadFlow() [][]string {
	file, err := os.Open(*netflows)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return [][]string{}
	}
	defer file.Close()

	reader := csv.NewReader(file)

	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error reading CSV:", err)
		return [][]string{}
	}

	return records
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

// log kid
func logKid(kid uint32) {
	file, err := os.OpenFile(*kidLog, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	info, _ := file.Stat()
	if info.Size() == 0 {
		writer.Write([]string{"kid"})
	}

	if err := writer.Write([]string{fmt.Sprintf("%d", kid)}); err != nil {
		panic(err)
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		panic(err)
	}
}

func main() {
	// parse command line arguments
	flag.Parse()

	// create blockchain connection
	conn_ledger, ledger := connectLedger()
	defer conn_ledger.Close()

	// create csp connection
	conn_csp, cs_provider := connectCsp()
	defer conn_csp.Close()

	// create context
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// prepare kid
	counter, kid := 0, uuid.New().ID()
	logKid(kid)

	// publish Garbled Circuit
	gc := &csp.GarbledCircuit{Kid: kid, F: []byte{0}, E: rand.Uint32()}
	cs_provider.PublishGC(ctx, gc)

	// load traffic
	flows := loadFlow()

	// generate traffic
	for _, flowrecord := range flows {
		// wait a while
		time.Sleep(time.Duration(rand.IntN(3)) * time.Second)

		// new kid after every 1000 flows
		if counter >= 1000 {
			kid = uuid.New().ID()
			logKid(kid)
		}
		counter++

		// create a random flow
		// flow := generateFlow(kid)
		flow := &blockchain.Flow{Id: uuid.New().ID(), Kid: kid, Destination: []byte(flowrecord[8]), Source: []byte(flowrecord[7])}

		// publish a the flow to the ledger
		err := publishFlow(ctx, ledger, flow)

		if err != nil {
			break
		}
	}
}
