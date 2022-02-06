#!/bin/bash

set -eu

PROTOC_LINK='https://github.com/protocolbuffers/protobuf/releases/download/v3.18.0/protoc-3.18.0-linux-x86_64.zip'
curl -Lo protoc.zip "$PROTOC_LINK"
unzip protoc.zip -d /tmp/protoc
mv /tmp/protoc/bin/protoc /usr/local/bin
mv /tmp/protoc/include/* /usr/local/include
rm protoc.zip
