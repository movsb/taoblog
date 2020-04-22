api: _protos _server _client _theme

_protos:
	./scripts/build-protos.sh

_server:
	go build -o taoblog ./server/

_client:
	go build -o ./client/taoblog ./client/
_theme:
	@cd themes/blog/statics/sass && ./make_style.sh

.PHONY: _protos _server _client _theme
