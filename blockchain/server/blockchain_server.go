// This module implements a basic blockchain simulation

package main

import (
	"context"
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"os"
	"sync"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/mboom/MedCTI/blockchain/proto"
)

var (
	ledger string
	host = flag.String("host", "localhost", "The hostname or IP address that will be used to listen.")
	port = flag.Int("port", 50051, "The server port")
)

type blockchainServer struct {
	pb.UnimplementedBlockchainServer
	mu sync.Mutex
}

func (bs *blockchainServer) PublishLogData(_ context.Context, flow *pb.Flow) (*empty.Empty, error) {
	// prevent opening the file multiple times
	bs.mu.Lock()

	// open the file that contains the simulated ledger
	f, err := os.OpenFile(ledger, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	defer bs.mu.Unlock()

	// convert a flow message to a message byte stream
	data, err := proto.Marshal(flow)
	if err != nil {
		panic(err)
	}

	// calculate the size of the message byte stream
	size := make([]byte, 4)
	_, err = binary.Encode(size, binary.BigEndian, uint32(len(data)))
	if err != nil {
		panic(err)
	}

	// insert the 4-byte size before the message byte stream
	data = append(size, data...)

	// write the data to the ledger file
	if _, err = f.Write(data); err != nil {
		panic(err)
	}

	return &empty.Empty{}, nil
}

func (bs *blockchainServer) FetchLogData(_ *empty.Empty, stream pb.Blockchain_FetchLogDataServer) error {
	// prevent opening the file multiple times
	bs.mu.Lock()

	// open the file that contains the simulated ledger
	f, err := os.OpenFile(ledger, os.O_RDONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	defer bs.mu.Unlock()

	// read the ledger file until the end of the file
	for err == nil {
		// read the 4-byte size of the next message
		sizebuffer := make([]byte, 4)
		_, err = f.Read(sizebuffer)
		if err != nil {
			panic(err)
		}

		// convert the size from byte array to uint32
		size := binary.BigEndian.Uint32(sizebuffer)

		// read the next message bytes
		data := make([]byte, size)
		_, err = f.Read(data)
		if err != nil {
			panic(err)
		}

		// convert the message byte stream to a flow message
		flow := &pb.Flow{}
		err = proto.Unmarshal(data, flow)
		if err != nil {
			panic(err)
		}

		// send the flow to the receiver
		if err := stream.Send(flow); err != nil {
			panic(err)
		}
	}
	return nil
}

func main() {
	// parse command line arguments
	flag.StringVar(&ledger, "ledger_file", "/var/tmp/blockchain-ledger.binpb", "The file where simulated ledger data is stored.")
	flag.Parse()

	// prepare listener socket for the RPC server
	lis, err := net.Listen("tcp", fmt.Sprintf("%v:%d", *host, *port))
	if err != nil {
		panic(err)
	}

	// create and start blockchain RPC server
	gRPCServer := grpc.NewServer()
	pb.RegisterBlockchainServer(gRPCServer, &blockchainServer{})
	gRPCServer.Serve(lis)
}