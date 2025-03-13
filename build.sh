#!/bin/bash
set -e

echo "Building UI..."
pushd ui
pnpm run build
popd

echo "Copying UI build to embed directory..."
rm -rf internal/static/dist
cp -r ui/dist internal/static/

echo "Building Go binary..."
go build -o srun cmd/srun/main.go
