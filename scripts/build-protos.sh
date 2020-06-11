#!/bin/bash

set -euo pipefail

protoc \
	-I. \
	-I/usr/local/include \
	-I"$GOPATH/pkg/mod" \
	-I"$GOPATH/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.8.5/third_party/googleapis" \
	-I"$GOPATH/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.8.5" \
	--go_out=plugins=grpc,paths=source_relative:. \
	protocols/backup.proto \
	protocols/service.proto \
	protocols/comment.proto \
	protocols/post.proto
