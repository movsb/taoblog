# Makefile for taoblog

OBJS=
OBJS+=admin-scripts
OBJS+=-client
OBJS+=plugins-codemirror
OBJS+=plugins-highlight
OBJS+=-server
OBJS+=-theme

all: $(OBJS)

admin-scripts:
	cd admin/scripts  && ./get-scripts.sh

-client:
	cd client && go build

plugins-codemirror:
	cd plugins/codemirror && ./get.sh

plugins-highlight:
	cd plugins/highlight && ./make_style.sh

-server:
	cd server && go build

-theme:
	cd theme/sass && ./make_style.sh
