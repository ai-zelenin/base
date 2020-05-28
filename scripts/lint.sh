#!/usr/bin/env bash

export PATH=$PATH:${GOPATH}/bin

if [[ ! -f ${GOPATH}/bin/golangci-lint ]]; then
  GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.27.0
  go get -u golang.org/x/tools/cmd/goimports
fi

golangci-lint -c res/.golangci.yml run ./cmd/... ./libs/... ./componets/...
