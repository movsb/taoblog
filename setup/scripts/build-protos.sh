#!/bin/bash

set -eu

GOPATH=${GOPATH:-$(go env GOPATH)}

cd protocols

protoc \
	-I. \
	-I/usr/local/include \
	-I"$GOPATH/pkg/mod" \
	-I"$GOPATH/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.16.0/third_party/googleapis" \
	-I"$GOPATH/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.16.0" \
	--go_out=paths=source_relative:go/proto \
	--go-grpc_out=paths=source_relative:go/proto \
	--grpc-gateway_out=paths=source_relative:go/proto \
	--swagger_out=allow_merge=true,merge_file_name="docs/taoblog",logtostderr=true:. \
	--swift_out=swift \
	--grpc-swift_out=Visibility=Internal,Server=false,Client=true,TestClient=false:swift \
	backup.proto \
	service.proto \
	comment.proto \
	post.proto \
	search.proto \
	config.proto
