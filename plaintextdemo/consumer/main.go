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

func TapTraffic() blockchain.Flow {
	flow := blockchain.Flow{id: uuid.New(), kid: uuid.New(), destination: [1]byte{2}, source: [1]byte{1} }
	return flow
}
