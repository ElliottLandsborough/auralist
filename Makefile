build:
	GOOS=linux GOARCH=amd64 go build -o bin/auralist *.go

run:
	go run *.go

index:
	go run *.go index

parsemp3:
	go run *.go parse:mp3

deps:
	go get ./...

test:
	go get github.com/stretchr/testify/assert
	go test -v ./... -race -coverprofile=coverage.txt -covermode=atomic

clean:
	rm bin/auralist

ssh:
	docker exec -ti auralist /bin/bash

devices:
	docker exec -ti auralist aplay -l

kill:
	docker kill auralist

restart:
	docker restart auralist
