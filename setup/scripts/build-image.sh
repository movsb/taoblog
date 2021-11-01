#!/bin/bash

set -eu

_run_in_xgo() {
	docker run -i --rm \
		-v "$(pwd)":/taoblog \
		-v "$(go env GOPATH)":/go \
		-e GOPATH=/go \
		-e USER="$USER" \
		-e HOSTNAME="$(hostname)" \
		-e DATE="$(date +'%F %T %z')" \
		-w /taoblog \
		--entrypoint bash \
		karalabe/xgo-latest \
		"$@"
}

(cd themes/blog/statics/sass && ./make_style.sh)

mkdir -p docker/setup/data
rsync -aPvh --delete setup/data/ docker/setup/data/

mkdir -p docker/themes/blog
rsync -aPvh --delete themes/blog/{statics,templates} docker/themes/blog/

mkdir -p docker/admin
rsync -aPvh --delete admin/login.html docker/admin

mkdir -p docker/protocols/docs
rsync -aPvh --delete protocols/docs docker/protocols

_run_in_xgo -c 'CGO_ENABLED=1 GOOS=linux GOARCH=amd64 ./setup/scripts/cross-build.sh'
IMAGE=taocker/taoblog:amd64-latest
(cd docker && docker build -t $IMAGE -f Dockerfile-amd64 .)

#_run_in_xgo -c 'CC=arm-linux-gnueabi-gcc-5 CGO_ENABLED=1 GOOS=linux GOARCH=arm ./setup/scripts/cross-build.sh'
#IMAGE=taocker/taoblog:arm-latest
#(cd docker && docker build -t $IMAGE -f Dockerfile-arm .)
