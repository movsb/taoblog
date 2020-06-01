#!/bin/bash

set -euo pipefail

docker run -it -v "$(pwd)":/taoblog -w /taoblog -v "$(go env GOPATH)":/go -e GOPATH=/go --entrypoint bash karalabe/xgo-latest -c 'CGO_ENABLED=1 go build -v -o docker/taoblog ./server/'

(cd themes/blog/statics/sass && ./make_style.sh)

rsync -aPvh --delete setup/data/ docker/setup/data/

mkdir -p docker/themes/blog
rsync -aPvh --delete themes/blog/{statics,templates} docker/themes/blog/

mkdir -p docker/admin
rsync -aPvh --delete admin/{statics,templates} docker/admin

IMAGE=taocker/taoblog:latest
(cd docker && docker build -t $IMAGE .)
