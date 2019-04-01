#!/bin/bash

[ -d aes2htm ] || git clone https://github.com/movsb/aes2htm.git || exit 1

cd aes2htm && git pull origin master && go build && mv aes2htm ../bin

