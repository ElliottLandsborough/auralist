build:
	GOOS=linux GOARCH=amd64 go build -o bin/auralist *.go

.PHONY: api
api:
	GOOS=linux GOARCH=amd64 go build -o bin/auralist-api api/*.go

listen:
	go run *.go listen

collect:
	go run *.go collectPaths

process:
	go run *.go processPaths

tag:
	go run *.go parsetags

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
