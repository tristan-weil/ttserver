#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

SUPER_LINTER_VERSION=v3.13.5
GOLANGCI_LINT_VERSION=v1.32.2
WHAT=$1

if [ "${WHAT}" = "super-linter" -o -z "${WHAT}" ]; then
  echo "Starting super-linter..."
  docker run -v "$(pwd)":/tmp/lint -e RUN_LOCAL=true -e LOG_FILE=/dev/null -e VALIDATE_GO=false github/super-linter:${SUPER_LINTER_VERSION}
fi

if [ "${WHAT}" = "golangci" -o -z "${WHAT}"  ]; then
  echo "Starting golangci-lint..."
  docker run -v "$(pwd)":/app -w /app golangci/golangci-lint:${GOLANGCI_LINT_VERSION} golangci-lint -c /app/.golangci.yml run
fi
