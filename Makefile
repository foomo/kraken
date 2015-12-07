SHELL := /bin/bash

clean:
	rm -f bin/kraken-linux
	rm -f bin/kraken-darwin
build:
	GOOS=linux GOARCH=amd64 go build -o bin/kraken-linux kraken-server/kraken-server.go
	GOOS=darwin GOARCH=amd64 go build -o bin/kraken-darwin kraken-server/kraken-server.go
test:
	go test github.com/foomo/kraken
