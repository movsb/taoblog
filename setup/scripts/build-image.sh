#!/bin/bash

set -euo pipefail

_run_in_xgo() {
	docker run -it --rm \
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

CGO_ENABLED=1 GOOS=linux GOARCH=amd64 _run_in_xgo ./setup/scripts/cross-build.sh

(cd themes/blog/statics/sass && ./make_style.sh)

rsync -aPvh --delete setup/data/ docker/setup/data/

mkdir -p docker/themes/blog
rsync -aPvh --delete themes/blog/{statics,templates} docker/themes/blog/

mkdir -p docker/admin
rsync -aPvh --delete admin/{statics,templates} docker/admin

mkdir -p docker/protocols/docs
rsync -aPvh --delete protocols/docs docker/protocols

IMAGE=taocker/taoblog:latest
(cd docker && docker build -t $IMAGE .)
