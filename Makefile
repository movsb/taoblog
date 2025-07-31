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

.PHONY: blocknote
blocknote:
	cd admin/blocknote && npm install && npm run build
	cd admin && rm -rf statics/blocknote && mkdir -p statics/blocknote && \
		cp -r blocknote/dist/assets statics/blocknote && \
		cp blocknote/dist/blocknote.* statics/blocknote && \
		cd statics/blocknote && sed -E 's/\/assets\//assets\//g' blocknote.css > a.css && \
			mv a.css blocknote.css

