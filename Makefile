.PHONY: all build test cross-arm64 cross-amd64 clean install

BINARY_NAME=docroid
BIN_DIR=bin

all: build

build:
	@mkdir -p $(BIN_DIR)
	go build -ldflags="-s -w" -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/docroid
	@echo "Built $(BIN_DIR)/$(BINARY_NAME)"

test:
	go test -v ./...

cross-arm64:
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o $(BIN_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/docroid
	@echo "Built static $(BIN_DIR)/$(BINARY_NAME)-linux-arm64 (Android / Termux)"

cross-amd64:
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BIN_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/docroid
	@echo "Built static $(BIN_DIR)/$(BINARY_NAME)-linux-amd64"

install: build
	@if [ -n "$$PREFIX" ]; then \
		cp $(BIN_DIR)/$(BINARY_NAME) $$PREFIX/bin/$(BINARY_NAME); \
		chmod +x $$PREFIX/bin/$(BINARY_NAME); \
		echo "Installed to $$PREFIX/bin/$(BINARY_NAME)"; \
	else \
		cp $(BIN_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME); \
		echo "Installed to /usr/local/bin/$(BINARY_NAME)"; \
	fi

clean:
	rm -rf $(BIN_DIR)
