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

.PHONY: compose
compose:
	mkdir -p run
	[ -f run/docker-compose.yml ] || cp setup/compose/docker-compose.yml run
	[ -f run/taoblog.db ] || touch run/taoblog.db
	[ -d run/prometheus ] || cp -R setup/compose/prometheus run
	[ -d run/grafana ] || cp -R setup/compose/grafana run
	(cd run && docker-compose up)

.PHONY: tools
tools:
	go install \
		github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway \
		github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger \
		github.com/golang/protobuf/protoc-gen-go
