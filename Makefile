SHELL := /bin/bash

#TAG=`git describe --exact-match --tags $(git log -n1 --pretty='%h') 2>/dev/null || git rev-parse --abbrev-ref HEAD`

all: build test
clean:
	rm -f .image_id
	rm -fv bin/krake*
build: clean
	go build -o bin/kraken kraken-server/kraken-server.go
build-arch: clean build-linux
	GOOS=darwin GOARCH=amd64 go build -o bin/kraken-darwin-amd64 kraken-server/kraken-server.go
build-linux: clean
	GOOS=linux GOARCH=amd64 go build -o bin/kraken-linux-amd64 kraken-server/kraken-server.go
build-docker: clean build-arch prepare-docker
	docker build -t foomo/kraken:latest .
prepare-docker:
	curl -o docker/files/cacert.pem https://curl.haxx.se/ca/cacert.pem
release: clean build-linux prepare-docker
	git add -f docker/files/cacert.pem
	git add -f bin/kraken-linux-amd64
	git commit -m 'build release candidate - new binary added for docker autobuild'
	# please make sure that version number has been bumped, then tag and push the git repo
test:
	go test ./...