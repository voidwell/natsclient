#!/bin/bash

apt install -y libprotoc-dev libprotobuf-dev protobuf-compiler

export GOPATH=$HOME/go
export GOROOT=/usr/local/go
export PATH=$PATH:/$GOROOT/bin:/$GOPATH/bin

go get -u google.golang.org/grpc
go get -u github.com/golang/protobuf/protoc-gen-go

find ./transport -name '*.proto' -type f | while read line; do
    protoc --go_out=paths=source_relative:. $line
done