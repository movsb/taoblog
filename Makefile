.PHONY: protos
protos:
	./setup/scripts/build-protos.sh

.PHONY: test
test:
	go test ./...

.PHONY: generate
generate:
	go generate ./...

.PHONY: build
build:
	./setup/scripts/cross-build.sh

.PHONY: build-image
build-image: 
	./setup/scripts/build-image.sh

.PHONY: push-image
push-image:
	docker push taocker/taoblog:amd64-latest

.PHONY: tools
tools:
	go install \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
		google.golang.org/protobuf/cmd/protoc-gen-go \
		google.golang.org/grpc/cmd/protoc-gen-go-grpc \
		github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
