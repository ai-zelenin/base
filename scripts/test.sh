#!/usr/bin/env bash
go test -v -gcflags=-l ./components/...
go test -v -gcflags=-l ./libs/...
