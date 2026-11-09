BIN := bin/procwatch

.PHONY: build test lint run health clean

build:
	go build -o $(BIN) ./cmd/procwatch

test:
	go test ./...

lint:
	gofmt -l .
	go vet ./...
	shellcheck scripts/*.sh

run: build
	./$(BIN) snapshot

health: build
	scripts/health.sh $(BIN)

clean:
	rm -rf bin
