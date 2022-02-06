#!/bin/bash

set -eu

OUTPUT='docker/taoblog'

builtOn="${USER:-root}@${HOSTNAME:-$(hostname)}"
builtAt="${DATE:-$(date +'%F %T %z')}"
goVersion=$(go version | sed 's/go version //')
gitAuthor=$(git show -s --format='format:%aN <%ae>' HEAD)
gitCommit=$(git rev-parse --short HEAD)

ldflags="\
-X 'github.com/movsb/taoblog/modules/version.BuiltOn=$builtOn' \
-X 'github.com/movsb/taoblog/modules/version.BuiltAt=$builtAt' \
-X 'github.com/movsb/taoblog/modules/version.GoVersion=$goVersion' \
-X 'github.com/movsb/taoblog/modules/version.GitAuthor=$gitAuthor' \
-X 'github.com/movsb/taoblog/modules/version.GitCommit=$gitCommit' \
"

# 静态链接、SQLite 与 CGO
# https://www.arp242.net/static-go.html
ldflags="$ldflags -extldflags=-static"
tags="osusergo,netgo,sqlite_omit_load_extension"

go build -ldflags "$ldflags" -tags "$tags" -v -o "$OUTPUT"
