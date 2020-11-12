#!/bin/bash

set -o errexit
set -o pipefail

if [ -z "${VERSION}" ]; then
  VERSION=$(git rev-parse --short HEAD)
fi

if [ -z "$DATE" ]; then
  DATE=$(date -u '+%Y-%m-%dT%H:%M:%S')
fi

GIT_REPO_URL='github.com/tristan-weil/ttserver/version'
GO_BUILD_CMD="go build -ldflags"
GO_BUILD_OPT="-s -w -X ${GIT_REPO_URL}.Version=${VERSION} -X ${GIT_REPO_URL}.BuildDate=${DATE}"
BUILD_DIR="build"

echo "Building Artifact..."
CGO_ENABLED=0 ${GO_BUILD_CMD} "${GO_BUILD_OPT}" -o "${BUILD_DIR}/ttserver"

echo "Building Docker..."
docker build --pull -t latest .

echo "Starting Docker..."
docker run -it --rm --name ttserver -p 7070:7272 -e TTSERVER_LISTENER_ADDRESS=0.0.0.0:7272 latest
