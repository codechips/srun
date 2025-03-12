#!/bin/bash
set -e

echo "Building UI..."
cd ui
npm run build
cd ..

echo "Copying UI build to embed directory..."
mkdir -p internal/static
rm -rf internal/static/*
cp -r ui/dist/* internal/static/

echo "Building Go binary..."
go build -o srun cmd/srun/main.go
