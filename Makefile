LINT := bin/golangci-lint

install: $(LINT)
	go mod download

$(LINT):
	curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b bin/ v2.15.0

build:
	go build -o bin/server ./cmd/server
	go build -o bin/cli ./cmd/cli
	go build -o bin/gui ./cmd/gui

run-server:
	go run ./cmd/server

run-client:
	go run ./cmd/cli

run-client-gui:
	go run ./cmd/gui

lint: $(LINT)
	$(LINT) run ./...

fmt: $(LINT)
	$(LINT) fmt ./...

test:
	go test -race ./...

clean:
	rm -rf bin
	go clean

.PHONY: install build run-server run-client run-client-gui lint test clean
