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
VERSION=$(git describe --tags --always --dirty)
COMMIT=$(git rev-parse --short HEAD)
BUILD_DATE=$(date -u '+%Y-%m-%d_%H:%M:%S')

go build -o srun \
  -ldflags "-X srun/internal/version.Version=$VERSION \
            -X srun/internal/version.GitCommit=$COMMIT \
            -X srun/internal/version.BuildDate=$BUILD_DATE" \
  cmd/srun/main.go
