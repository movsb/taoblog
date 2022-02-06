#!/bin/bash

set -eu

SASS_LINK='https://github.com/sass/dart-sass/releases/download/1.49.7/dart-sass-1.49.7-linux-x64.tar.gz'
curl -Lo sass.tgz "$SASS_LINK"
mkdir -p /tmp/sass && tar xzvf sass.tgz -C /tmp/sass
mv /tmp/sass/dart-sass/sass /usr/local/bin
