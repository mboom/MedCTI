package blockchain

import (
	"context"
	"encoding/binary"
	"os"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/mboom/MedCTI/blockchain"
)

const ledger = "/var/tmp/blockchain-ledger.binpb"

type blockchainServer struct {
	pb.UnimplementedBlockchainServer
}

func (bs *blockchainServer) PublishLogData(_ context.Context, flow *pb.Flow) (*empty.Empty, error) {
	f, err := os.OpenFile(ledger, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	data, err := proto.Marshal(flow)
	if err != nil {
		panic(err)
	}
	size := make([]byte, 4)
	_, err = binary.Encode(size, binary.BigEndian, uint32(len(data)))
	if err != nil {
		panic(err)
	}
	data = size + data

	if _, err = f.Write(data); err != nil {
		panic(err)
	}

	return empty.Empty{}, nil
}

func (bs *blockchainServer) FetchLogData(_ *empty.Empty, stream Blockchain_FetchLogDataClient) error {
	f, err := os.OpenFile(ledger, os.O_RDONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	for err == nil {
		sizebuffer := make([]byte, 4)
		_, err = f.Read(sizebuffer)
		if err != nil {
			panic(err)
		}
		size := binary.BigEndian.Uint32(sizebuffer)

		data := make([]byte, size)
		_, err = f.Read(data)
		if err != nil {
			panic(err)
		}

		flow := &pb.Flow{}
		err = proto.Unmarshal(data, flow)
		if err != nil {
			panic(err)
		}

		if err := stream.Send(flow); err != nil {
			panic(err)
		}
	}
	return nil
}