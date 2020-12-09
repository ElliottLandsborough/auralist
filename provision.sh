# unfinished
docker run --name auralist -t -d ubuntu:latest
docker exec -ti auralist /bin/bash
export DEBIAN_FRONTEND=noninteractive
apt update
apt upgrade -y
apt install -y git wget make

cd
wget https://golang.org/dl/go1.15.6.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.15.6.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
export GOPATH=~/go
go version

go get -v -u github.com/ElliottLandsborough/auralist
cd $GOPATH/src/github.com/ElliottLandsborough/auralist
make deps
nano config.yml
make collect
make files