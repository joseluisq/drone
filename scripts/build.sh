#!/bin/sh

echo "building docker images for ${GOOS}/${GOARCH} ..."

REPO="github.com/joseluisq/drone"

# compile the server using the cgo
go version
go build -v \
    -ldflags "-s -w \
        -X 'main.version=0.0.0' \
        -X 'main.buildTime=$BUILD_TIME' \
        -X 'main.buildCommit=$BUILD_VCS_REF'" \
    -a -o release/linux/${GOARCH}/drone-server ${REPO}/cmd/drone-server
ls -ogh release/linux/${GOARCH}/drone-server

# # compile the runners with gcc disabled
# export CGO_ENABLED=0
# go build -o release/linux/${GOARCH}/drone-agent      ${REPO}/cmd/drone-agent
# go build -o release/linux/${GOARCH}/drone-controller ${REPO}/cmd/drone-controller
