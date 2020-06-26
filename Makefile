.PHONY: all
all: protos_ theme_ main_

.PHONY: protos_
protos_:
	./scripts/build-protos.sh

.PHONY: theme_
theme_:
	@cd themes/blog/statics/sass && ./make_style.sh

.PHONY: main_
main_:
	./scripts/cross-build.sh
	cp docker/taoblog .

.PHONY: build-image
build-image: 
	./scripts/build-image.sh

.PHONY: push-image
push-image:
	docker push taocker/taoblog:latest

.PHONY: tools
tools:
	go install \
		github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway \
		github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger \
		github.com/golang/protobuf/protoc-gen-go
