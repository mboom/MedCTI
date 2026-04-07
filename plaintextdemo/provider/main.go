package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	blockchain "github.com/mboom/MedCTI/blockchain/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	host  = flag.String("host", "localhost", "The hostname or IP address that will be used to listen.")
	port  = flag.Int("port", 50051, "The server port")
	intel = flag.String("intel", "../data/threats-fs1000-LITNET-2020.csv", "File with collected threat intelligence.")
	match = flag.String("match", "../data/matches.csv", "File to write found threats.")
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

func logMatch(flow blockchain.Flow) {
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

		dest := cti(fmt.Sprintf("%d", flow.Destination))
		src := cti(fmt.Sprintf("%d", flow.Source))

		if dest || src {
			logMatch(*flow)
		}

		// print data
		// fmt.Println(flow)
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
