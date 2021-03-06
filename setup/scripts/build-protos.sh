#!/bin/bash

set -euo pipefail

protoc \
	-I. \
	-I/usr/local/include \
	-I"$GOPATH/pkg/mod" \
	-I"$GOPATH/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.14.6/third_party/googleapis" \
	-I"$GOPATH/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.14.6" \
	--go_out=plugins=grpc,paths=source_relative:. \
	--grpc-gateway_out=logtostderr=true,paths=source_relative:. \
	--swagger_out=allow_merge=true,merge_file_name="protocols/docs/taoblog",logtostderr=true:. \
	protocols/backup.proto \
	protocols/service.proto \
	protocols/comment.proto \
	protocols/post.proto
