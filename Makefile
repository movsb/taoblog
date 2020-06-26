all: protos_ main_ theme_

.PHONY: protos_ main_ theme_

protos_:
	./scripts/build-protos.sh

main_:
	./scripts/cross-build.sh
	cp docker/taoblog .

theme_:
	@cd themes/blog/statics/sass && ./make_style.sh

.PHONY: build-image
build-image: 
	./scripts/build-image.sh

.PHONY: push-image
push-image:
	docker push taocker/taoblog:latest
