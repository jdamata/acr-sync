export GO111MODULE=on
VERSION=$(shell git describe --tags --candidates=1 --dirty)
BUILD_FLAGS=-ldflags="-X main.version=$(VERSION)"
SRC=$(shell find . -name '*.go')

.PHONY: all clean release install

all: acr-sync-linux-amd64 acr-sync-darwin-amd64

clean:
	rm -f acr-sync acr-sync-linux-amd64 acr-push-darwin-amd64

acr-sync-linux-amd64: $(SRC)
	GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o $@ .

acr-sync-darwin-amd64: $(SRC)
	GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) -o $@ .

install:
	rm -f acrsync
	go build $(BUILD_FLAGS) .
	mv acr-sync ~/bin/
