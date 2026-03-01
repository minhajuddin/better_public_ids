.PHONY: build test test-race test-v vet fuzz fuzz-parse fuzz-text fuzz-json bench clean all

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

fuzz: fuzz-parse fuzz-text fuzz-json

fuzz-parse:
	go test ./... -fuzz=FuzzParse -fuzztime=$(FUZZTIME)

fuzz-text:
	go test ./... -fuzz=FuzzUnmarshalText -fuzztime=$(FUZZTIME)

fuzz-json:
	go test ./... -fuzz=FuzzUnmarshalJSON -fuzztime=$(FUZZTIME)

bench:
	go test ./... -bench=. -benchmem

clean:
	go clean -testcache -fuzzcache
