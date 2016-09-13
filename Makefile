SHELL := /bin/bash

TAG=`git describe --exact-match --tags $(git log -n1 --pretty='%h') 2>/dev/null || git rev-parse --abbrev-ref HEAD`

all: build test
clean:
	rm -f .image_id
	rm -fv bin/krake*
build: clean
	go build -o bin/kraken kraken-server/kraken-server.go
build-arch: clean
	GOOS=linux GOARCH=amd64 go build -o bin/kraken-linux-amd64 kraken-server/kraken-server.go
	GOOS=darwin GOARCH=amd64 go build -o bin/kraken-darwin-amd64 kraken-server/kraken-server.go
build-docker: clean build-arch
	curl -o docker/files/cacert.pem https://curl.haxx.se/ca/cacert.pem
	docker build -q -t docker-registry.bestbytes.net/kraken:$(TAG) . > .image_id
	docker tag `cat .image_id` docker-registry.bestbytes.net/kraken:latest
	echo "# tagged container `cat .image_id` as docker-registry.bestbytes.net/kraken:$(TAG)"
	rm -f .image_id
test:
	go test ./...
