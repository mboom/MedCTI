Demo Plain Text
===============

Installation
------------

```
setupfile="go1.24.3.linux-amd64.tar.gz"
wget "https://go.dev/dl/$setupfile"

sudo rm -r /usr/local/lib/go
sudo tar -C /usr/local/lib -xzf $setupfile

sudo ln -s /usr/local/lib/go/bin/go /usr/local/bin/go
sudo ln -s /usr/local/lib/go/bin/gofmt /usr/local/bin/gofmt

rm $setupfile

git clone git@github.com:mboom/MedCTI.git
```

Execution
---------
```
cd MedCTI/DemoPlainText/MedCTI
go run .
```