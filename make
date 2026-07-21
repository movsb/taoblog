#!/bin/bash

protos() {
	./setup/scripts/build-protos.sh
}

test() {
	go test ./...
}

cover() {
	dir="$$TMPDIR"
	go test -coverpkg=./... -coverprofile="$$dir"/coverage.out ./...
	go tool cover -html="$$dir"/coverage.out
}

generate() {
	go generate ./...
}

build() {
	./setup/scripts/cross-build.sh
}

build-image() {
	./setup/scripts/build-image.sh
}

push-image() {
	docker push taocker/taoblog:amd64-latest
}

tools() {
	go install \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
		google.golang.org/protobuf/cmd/protoc-gen-go \
		google.golang.org/grpc/cmd/protoc-gen-go-grpc \
		github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
}

for cmd in "${@}"; do
	$cmd
done
