package main

import (
	"fmt"
	"github.com/google/uuid"
	blockchain "github.com/mboom/MedCTI/blockchain/proto"
)

func main() {
	name := "MedCTI"
	fmt.Println("Project for", name)
	fmt.Println(TapTraffic())
}

func TapTraffic() *blockchain.Flow {
	flow := blockchain.Flow{Id: uuid.New().ID(), Kid: uuid.New().ID(), Destination: []byte{2}, Source: []byte{1} }
	return &flow
}
