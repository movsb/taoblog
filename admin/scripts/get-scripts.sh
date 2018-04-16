#!/bin/bash

[ -f marked.js ] && exit 0

wget -O marked.js https://github.com/markedjs/marked/raw/master/marked.min.js
