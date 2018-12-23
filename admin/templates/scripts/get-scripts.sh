#!/bin/bash

[ -f marked.js ] && exit 0

curl -L https://github.com/markedjs/marked/raw/master/marked.min.js > marked.js

