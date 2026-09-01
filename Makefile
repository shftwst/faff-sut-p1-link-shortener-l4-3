# Engineering gates for the link-shortener api. The toolchain runs inside the
# pinned Go image so the checks are reproducible and need no host Go install.
GO_IMAGE ?= golang:1.22-alpine
GO_RUN   := docker run --rm -v "$(CURDIR)":/src -w /src -e GOTOOLCHAIN=local $(GO_IMAGE)

.PHONY: fmt vet build test tidy

## fmt: fail if any Go source is not gofmt-clean
fmt:
	$(GO_RUN) sh -c 'test -z "$$(gofmt -l .)" || { echo "unformatted:"; gofmt -l .; exit 1; }'

## vet: static analysis
vet:
	$(GO_RUN) go vet ./...

## build: compile all packages
build:
	$(GO_RUN) go build ./...

## test: run the unit test suite
test:
	$(GO_RUN) go test ./...

## tidy: sync go.mod/go.sum
tidy:
	$(GO_RUN) go mod tidy
