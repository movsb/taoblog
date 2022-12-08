#!/bin/bash

set -eu

docker run -i$([ -t 0 ] && echo t) --rm -v "$(pwd)/.tmp/go":/root/go -v "$(pwd)":/workspace -w /workspace --entrypoint bash -- taocker/taoblog-build-tool:latest  -c "$@"
