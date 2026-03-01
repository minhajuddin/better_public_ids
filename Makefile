.PHONY: build test test-race test-v vet fuzz bench clean all

all: vet build test

build:
	go build ./...

test:
	go test ./... -count=1

test-v:
	go test ./... -v -count=1

test-race:
	go test ./... -race -count=1

vet:
	go vet ./...

FUZZTIME ?= 10s

fuzz:
	go test ./... -fuzz=FuzzDeserialize -fuzztime=$(FUZZTIME)

bench:
	go test ./... -bench=. -benchmem

clean:
	go clean -testcache -fuzzcache
