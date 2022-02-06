#!/bin/bash

set -eu

mkdir -p docker/setup/data
rsync -aPvh --delete setup/data/ docker/setup/data/

mkdir -p docker/themes/blog
rsync -aPvh --delete themes/blog/{statics,templates} docker/themes/blog/

mkdir -p docker/admin
rsync -aPvh --delete admin/login.html docker/admin

mkdir -p docker/protocols/docs
rsync -aPvh --delete protocols/docs docker/protocols

cp taoblog docker/
IMAGE=taocker/taoblog:amd64-latest
(cd docker && docker build -t $IMAGE -f Dockerfile .)
