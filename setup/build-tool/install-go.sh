#!/bin/bash

set -eu

GO_LINK='https://go.dev/dl/go1.17.6.linux-amd64.tar.gz'
curl -Lo go.tgz "$GO_LINK"
tar xzvf go.tgz -C /usr/local
echo 'PATH="$PATH":/usr/local/go/bin:/root/go/bin' >> ~/.bashrc
rm go.tgz
