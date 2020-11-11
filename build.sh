#!/bin/bash

set -o errexit
set -o pipefail

if [ -z "${VERSION}" ]; then
  VERSION=$(git rev-parse --short HEAD)
fi

if [ -z "$DATE" ]; then
  DATE=$(date -u '+%Y-%m-%dT%H:%M:%S')
fi

echo "Building ${VERSION} ${DATE}"

GIT_REPO_URL='github.com/tristan-weil/ttserver/version'
GO_BUILD_CMD="go build -ldflags"
GO_BUILD_OPT="-s -w -X ${GIT_REPO_URL}.Version=${VERSION} -X ${GIT_REPO_URL}.BuildDate=${DATE}"
BUILD_DIR="build"

# Build amd64 binaries
OS_PLATFORM_ARG=(linux openbsd)
OS_ARCH_ARG=(amd64)
for OS in "${OS_PLATFORM_ARG[@]}"; do
  BIN_EXT=""
  if [ "$OS" == "windows" ]; then
    BIN_EXT=".exe"
  fi
  for ARCH in "${OS_ARCH_ARG[@]}"; do
    echo "Building binary for ${OS}/${ARCH}..."
    GOARCH=${ARCH} GOOS=${OS} CGO_ENABLED=0 ${GO_BUILD_CMD} "${GO_BUILD_OPT}" -o "${BUILD_DIR}/${VERSION}/${OS}/${ARCH}/ttserver${BIN_EXT}" .
    zip -j "${BUILD_DIR}/ttserver-${VERSION}-${OS}-${ARCH}.zip" "${BUILD_DIR}/${VERSION}/${OS}/${ARCH}/ttserver${BIN_EXT}"
  done
done
