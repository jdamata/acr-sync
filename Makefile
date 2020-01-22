export GO111MODULE=on
VERSION=$(shell git describe --tags --candidates=1 --dirty)
BUILD_FLAGS=-ldflags="-X main.version=$(VERSION)"
SRC=$(shell find . -name '*.go')

.PHONY: all clean release install

all: acrpush-linux-amd64 acrpush-darwin-amd64

clean:
	rm -f acrpush acrpush-linux-amd64 acrpush-darwin-amd64

k8vault-linux-amd64: $(SRC)
	GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o $@ .

k8vault-darwin-amd64: $(SRC)
	GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) -o $@ .

install:
	rm -f acrpush
	go build $(BUILD_FLAGS) .
	mv acrpush ~/bin/
