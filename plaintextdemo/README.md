Demo Plain Text
===============

Installation
------------

```bash
# Install Go
setupfile="go1.24.3.linux-amd64.tar.gz"
wget "https://go.dev/dl/$setupfile"

sudo rm -r /usr/local/lib/go
sudo tar -C /usr/local/lib -xzf $setupfile

sudo ln -s /usr/local/lib/go/bin/go /usr/local/bin/go
sudo ln -s /usr/local/lib/go/bin/gofmt /usr/local/bin/gofmt

rm $setupfile

# Clone repository
git clone git@github.com:mboom/MedCTI.git
cd MedCTI/plaintextdemo
```

Set and install dependencies

    go mod tidy

Execution
---------

Execute the cryptographic service provider

    go run ./csp

Execute the consumer

    go run ./consumer

Execute the provider

    go run ./provider