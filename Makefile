.PHONY: all
all: protos theme main

.PHONY: protos
protos:
	./setup/scripts/build-protos.sh

.PHONY: theme
theme:
	@cd theme/blog/styles && ./make_style.sh

.PHONY: main
main:
	./setup/scripts/cross-build.sh

.PHONY: build-image
build-image: 
	./setup/scripts/build-image.sh

.PHONY: push-image
push-image:
	docker push taocker/taoblog:amd64-latest
	#docker push taocker/taoblog:arm-latest
.PHONY: try
try:
	docker run -it --rm --name=taoblog -p 2564:2564 -p 2563:2563 taocker/taoblog:amd64-latest

.PHONY: tools
tools:
	go install \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
		google.golang.org/protobuf/cmd/protoc-gen-go \
		google.golang.org/grpc/cmd/protoc-gen-go-grpc \
		github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
