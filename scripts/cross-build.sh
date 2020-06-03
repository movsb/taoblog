#!/bin/bash

set -euo pipefail

builtOn="$USER@$HOSTNAME"
builtAt="$DATE"
goVersion=$(go version | sed 's/go version //')
gitAuthor=$(git show -s --format='format:%aN <%ae>' HEAD)
gitCommit=$(git rev-parse HEAD)

ldflags="\
-X 'main.builtOn=$builtOn' \
-X 'main.builtAt=$builtAt' \
-X 'main.goVersion=$goVersion' \
-X 'main.gitAuthor=$gitAuthor' \
-X 'main.gitCommit=$gitCommit' \
"

CGO_ENABLED=1 go build -ldflags "$ldflags" -v -o docker/taoblog
