package main

import (
	"context"
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/mboom/MedCTI/csp/plaintextdemo/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

var (
	vault string
	host  = flag.String("host", "localhost", "The hostname or IP address that will be used to listen.")
	port  = flag.Int("port", 50052, "The server port")
)

type cspServer struct {
	pb.UnimplementedCSPServer
	mu sync.Mutex
}

func (bs *cspServer) PublishGC(_ context.Context, circuit *pb.GarbledCircuit) (*empty.Empty, error) {
	// prevent opening the vault multiple times simultaneously
	bs.mu.Lock()

	// open the vault
	f, err := os.OpenFile(vault, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	defer bs.mu.Unlock()

	// convert a circuit to a message byte stream
	data, err := proto.Marshal(circuit)
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

	// write the data to the vault
	if _, err = f.Write(data); err != nil {
		panic(err)
	}

	return &empty.Empty{}, nil
}

func (bs *cspServer) FetchGC(_ context.Context, key *pb.Key) (*pb.GarbledCircuit, error) {
	// prevent opening the file multiple times simultaneously
	bs.mu.Lock()

	// open the vault
	f, err := os.OpenFile(vault, os.O_RDONLY|os.O_CREATE, 0600)
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
			break
		}

		// convert the size from byte array to uint32
		size := binary.BigEndian.Uint32(sizebuffer)

		// read the next message bytes
		data := make([]byte, size)
		_, err = f.Read(data)
		if err != nil {
			panic(err)
		}

		// convert the message byte stream to a Garbled Circuit
		circuit := &pb.GarbledCircuit{}
		err = proto.Unmarshal(data, circuit)
		if err != nil {
			panic(err)
		}

		if key.Kid == circuit.Kid {
			return circuit, nil
		}

	}

	return nil, nil
}

func main() {
	// parse command line arguments
	flag.StringVar(&vault, "ledger_file", "/var/tmp/csp-vault.binpb", "The file where simulated ledger data is stored.")
	flag.Parse()

	// prepare listener socket for the RPC server
	lis, err := net.Listen("tcp", fmt.Sprintf("%v:%d", *host, *port))
	if err != nil {
		panic(err)
	}

	// create and start blockchain RPC server
	gRPCServer := grpc.NewServer()
	pb.RegisterCSPServer(gRPCServer, &cspServer{})
	gRPCServer.Serve(lis)
}
