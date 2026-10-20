#!/bin/sh
set -e

echo "Building Docroid multi-arch binaries..."
mkdir -p bin

echo "1. Building native binary..."
go build -ldflags="-s -w" -o bin/docroid ./cmd/docroid

echo "2. Cross-compiling for Android ARM64 (aarch64)..."
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o bin/docroid-linux-arm64 ./cmd/docroid

echo "3. Cross-compiling for Linux AMD64 (x86_64)..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/docroid-linux-amd64 ./cmd/docroid

echo "4. Cross-compiling for Linux ARM (armv7)..."
CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -ldflags="-s -w" -o bin/docroid-linux-arm ./cmd/docroid

echo "Done! Built binaries in ./bin:"
ls -lh bin/
