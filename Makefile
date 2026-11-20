LINT := bin/golangci-lint

install: $(LINT)
	go mod download

$(LINT):
	curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b bin/ v2.15.0

build: build-server build-client build-client-gui

build-server:
	go build -o bin/server ./cmd/server

build-client:
	go build -o bin/cli ./cmd/cli

build-client-gui:
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

.PHONY: install build run-server run-client run-client-gui lint test clean \
	build-server build-cli build-gui
