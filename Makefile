SHELL := /bin/bash

clean:
	rm -vf bin/kraken
build:
	go build -o bin/kraken kraken-server/kraken-server.go
test:
	go test -v github.com/foomo/kraken
