#/bin/bash

curl -o codemirror.zip -L http://codemirror.net/codemirror.zip
unzip codemirror.zip
cd codemirror-*

cp lib/codemirror.css ..
cp lib/codemirror.js ..
cp mode/markdown/markdown.js ..
cp addon/dialog/dialog.css ..
cp addon/dialog/dialog.js ..
cp keymap/vim.js ..

cd ..
rm codemirror.zip
rm -rf codemirror-*

curl -o marked.min.js -L https://github.com/chjj/marked/raw/master/marked.min.js

