.PHONY: build
build:
	docker buildx build --platform linux/amd64 -t taocker/taoblog-build-tool:latest .
.PHONY: push
push:
	docker push taocker/taoblog-build-tool:latest
