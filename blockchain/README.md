The Blockchain package provides a simulation for a blockchain implementation. Hence, there is no real blockchain. There are only some remote procedure calls to simulate reading from and writing to a blockchain. This package is used to show how the MedCTI could be implemented on a blockchain.

Install dependencies:

```bash
# Install Go
setupfile="go1.24.3.linux-amd64.tar.gz"
wget "https://go.dev/dl/$setupfile"

sudo rm -r /usr/local/lib/go
sudo tar -C /usr/local/lib -xzf $setupfile

sudo ln -s /usr/local/lib/go/bin/go /usr/local/bin/go
sudo ln -s /usr/local/lib/go/bin/gofmt /usr/local/bin/gofmt

rm $setupfile

# Install Protocol Buffers compiler and gRPC compiler
sudo apt-get install protobuf-compiler
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Clone repository
git clone git@github.com:mboom/MedCTI.git
cd MedCTI/blockchain
```

Add GOPATH to your PATH variable to make the protocol compiler ```protoc``` locatable:

    export PATH="$PATH:$(go env GOPATH)/bin"

Compile the Blockchain service:

    protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/blockchain.proto

Set and install dependencies

    go mod tidy

Execute the server

    go run ./server