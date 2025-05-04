#!/bin/bash

set -eu

GO_LINK='https://go.dev/dl/go1.24.2.linux-amd64.tar.gz'
curl -Lo go.tgz "$GO_LINK"
tar xzvf go.tgz -C /usr/local
rm go.tgz
