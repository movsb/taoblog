api: _protos _server _client

_protos:
	./scripts/build-protos.sh

_server:
	go build -o taoblog ./server/

_client:
	go build -o ./client/taoblog ./client/

.PHONY: _protos _server _client
