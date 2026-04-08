.PHONY: build test lint clean integration-test

build:
	go build ./...

test:
	go test -race -count=1 ./...

integration-test:
	go test -race -count=1 -tags=integration ./...

lint:
	golangci-lint run ./...

clean:
	go clean -cache -testcache
