#!/bin/bash

set -euo pipefail

GOOS=linux GOARCH=amd64 go build -o ./docker/taoblog ./server/
(cd themes/blog/statics/sass && ./make_style.sh)

GOOS=linux GOARCH=amd64 go build -o ./docker/setup/init ./setup/init
rsync -aPvh --delete setup/data/ docker/setup/data/

mkdir -p docker/themes/blog
rsync -aPvh --delete themes/blog/{statics,templates,tools} docker/themes/blog/

mkdir -p docker/admin
rsync -aPvh --delete admin/{statics,templates} docker/admin

IMAGE=taocker/taoblog:latest
(cd docker && docker build -t $IMAGE .)
