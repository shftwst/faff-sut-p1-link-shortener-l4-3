#!/usr/bin/env bash
# Fail if any Go source is not gofmt-clean. Runs gofmt in the pinned Go image.
set -euo pipefail
unformatted=$(docker run --rm -v "$PWD":/src -w /src -e GOTOOLCHAIN=local golang:1.22-alpine gofmt -l .)
if [ -n "$unformatted" ]; then
  echo "gofmt: the following files are not formatted:"
  echo "$unformatted"
  exit 1
fi
echo "gofmt: clean"
