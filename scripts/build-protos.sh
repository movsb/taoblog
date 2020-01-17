#!/bin/bash

set -e

protoc --go_out=plugins=grpc,paths=source_relative:. protocols/backup.proto protocols/service.proto
