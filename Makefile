api: protos_ server_ client_ theme_

protos_:
	./scripts/build-protos.sh

server_:
	go build -o taoblog ./server/

client_:
	go build -o ./client/taoblog ./client/
theme_:
	@cd themes/blog/statics/sass && ./make_style.sh

.PHONY: protos_ server_ client_ theme_

.PHONY: build-image
build-image: 
	./scripts/build-image.sh
.PHONY: push-image
push-image:
	docker push taocker/taoblog:latest
