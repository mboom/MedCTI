package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"math"
	"net"
	"os"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	blockchain "github.com/mboom/MedCTI/blockchain/proto"
	"github.com/mboom/MedCTI/threatintel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	bc_host = flag.String("bcHost", "localhost", "The hostname or IP address that will be used to listen.")
	bc_port = flag.Int("bcPort", 50051, "The server port")
	ti_host = flag.String("tiHost", "localhost", "The hostname or IP address that will be used to share threat intel.")
	ti_port = flag.Int("tiPort", 50053, "The port that will be used to share threat intel.")
	intel   = flag.String("intel", "../data/threats-fs1000-LITNET-2020.csv", "File with collected threat intelligence.")
	match   = flag.String("match", "../data/matches.csv", "File to write found threats.")
)

type threatintelServer struct {
	threatintel.UnimplementedThreatIntelServer
	mu      sync.Mutex
	threats chan *blockchain.Flow
}

func (ti *threatintelServer) RequestThreatIntel(_ context.Context, keyId *threatintel.KeyId) (*threatintel.Indicators, error) {
	indicators := []byte{}
	count := 0
	fmt.Printf("Looking for threats for kid %d.\n", keyId.Id)
	for threat := range ti.threats {
		if threat.Kid == keyId.Id {
			fmt.Printf("Matched threat %d, with kid %d and key %d.\n", threat.Id, threat.Kid, keyId.Id)
			indicators = append(indicators, threat.Destination...)
			indicators = append(indicators, threat.Source...)
		} else if count == 20 {
			break
		} else {
			count++
		}
	}
	fmt.Printf("Sending %d threats\n", len(indicators))
	return &threatintel.Indicators{Addresses: indicators}, nil
}

// create connection to the blockchain simulator
func connectLedger() (*grpc.ClientConn, blockchain.BlockchainClient) {
	conn, err := grpc.NewClient(fmt.Sprintf("%v:%d", *bc_host, *bc_port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	ledger := blockchain.NewBlockchainClient(conn)
	return conn, ledger
}

// find threats
func cti(ioc string) bool {
	file, err := os.Open(*intel)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return false
	}
	defer file.Close()

	reader := csv.NewReader(file)

	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error reading CSV:", err)
		return false
	}

	for _, record := range records {
		if ioc == record[0] {
			return true
		}
	}

	return false
}

func logMatch(flow *blockchain.Flow) {
	file, err := os.OpenFile(*match, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	info, _ := file.Stat()
	if info.Size() == 0 {
		writer.Write([]string{"Id", "Kid", "Destination", "Source"})
	}

	if err := writer.Write([]string{fmt.Sprintf("%d", flow.Id), fmt.Sprintf("%d", flow.Kid), fmt.Sprintf("%d", flow.Destination), fmt.Sprintf("%d", flow.Source)}); err != nil {
		panic(err)
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		panic(err)
	}
}

func read(ctx context.Context, ledger blockchain.BlockchainClient, threats chan<- *blockchain.Flow) {
	var ledger_err error = nil
	for ledger_err == nil {
		// create blockchain stream reader
		data, ledger_err := ledger.FetchLogData(ctx, &empty.Empty{})
		if ledger_err != nil {
			panic(ledger_err)
		}

		// receive data from blockchain
		var err error = nil
		for err == nil {
			flow, err := data.Recv()
			if err != nil {
				break
			}
			// fmt.Printf("received flow %d\n", flow.Id)

			dest := cti(fmt.Sprintf("%d", flow.Destination[0]))
			src := cti(fmt.Sprintf("%d", flow.Source[0]))

			if dest || src {
				// logMatch(flow)
				threats <- flow
				fmt.Printf("found threat. s: %d, d: %d\n", flow.Source[0], flow.Destination[0])
			}

			// print data
			// fmt.Println(flow)
		}
	}
}

func main() {
	// parse command line arguments
	flag.Parse()

	// create connection
	conn_ledger, ledger := connectLedger()
	defer conn_ledger.Close()

	// create context
	ctx, cancel := context.WithTimeout(context.Background(), 24*time.Hour)
	defer cancel()

	// create Threats channel
	threats := make(chan *blockchain.Flow, math.MaxInt16)

	// read blockchain
	go read(ctx, ledger, threats)

	// prepare listener socket for the RPC server
	lis, err := net.Listen("tcp", fmt.Sprintf("%v:%d", *ti_host, *ti_port))
	if err != nil {
		panic(err)
	}

	// create and start Threat Intel RPC server
	gRPCServer := grpc.NewServer()
	threatintel.RegisterThreatIntelServer(gRPCServer, &threatintelServer{threats: threats})
	gRPCServer.Serve(lis)
}
