package blockchain

import (
	"context"
	"google.golang.org/grpc"
	empty "github.com/golang/protobuf/ptypes/empty"
	pb "github.com/mboom/MedCTI/blockchain"
)

type blockchainServer struct {
	pb.UnimplementedBlockchainServer
}

func (bs *blockchainServer) PublishLogData(context.Context, flow *pb.Flow) (*empty.Empty, error) {
	return nil
}

func (bs *blockchainServer) FetchLogData(*empty.Empty, stream Blockchain_FetchLogDataClient) error {
	return nil
}