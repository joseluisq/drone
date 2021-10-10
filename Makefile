BUILD_TIME ?= $(shell date -u '+%Y-%m-%dT%H:%m:%S')
BUILD_VCS_REF ?= $(shell git rev-parse --short HEAD)
REPO="github.com/joseluisq/drone"

install:
	@go version
	@go get -v golang.org/x/lint/golint
.PHONY: install

test:
	@go version
	@golint -set_exit_status ./...
	@go vet ./...
	@go test $$(go list ./... | grep -v /examples) \
		-v -timeout 30s -race -coverprofile=coverage.txt -covermode=atomic
.PHONY: test

build:
	@go version
	@go build -v \
		-ldflags "-s -w \
			-X 'main.version=0.0.0' \
			-X 'main.buildTime=$(BUILD_TIME)' \
			-X 'main.buildCommit=$(BUILD_VCS_REF)'" \
		-a -o release/linux/${GOARCH}/drone-server ${REPO}/cmd/drone-server
	@ls -ogh release/linux/${GOARCH}/drone-server
.PHONY: build
