#!/bin/bash

set -euo pipefail

builtOn="$USER@${HOSTNAME:-$(hostname)}"
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

CGO_ENABLED=1 go build -ldflags "$ldflags" -v -o docker/taoblog
