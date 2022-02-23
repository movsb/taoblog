#!/bin/bash

set -eu

IMAGE=taocker/taoblog:amd64-latest
docker build -t $IMAGE -f Dockerfile .
