#!/bin/bash

set -euo pipefail

(cd "$GOPATH"/src/github.com/movsb/taoblog && xgo --targets=linux/amd64 -dest docker ./server)

(cd themes/blog/statics/sass && ./make_style.sh)

rsync -aPvh --delete setup/data/ docker/setup/data/

mkdir -p docker/themes/blog
rsync -aPvh --delete themes/blog/{statics,templates,tools} docker/themes/blog/

mkdir -p docker/admin
rsync -aPvh --delete admin/{statics,templates} docker/admin

IMAGE=taocker/taoblog:latest
(cd docker && docker build -t $IMAGE .)
