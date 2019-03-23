#!/bin/bash

if ! hash sass 2>/dev/null; then
    echo "sass does not exist.";
    exit 1;
fi

if [ "$1" == "--watch" ]; then
    sass --watch style.scss:../style.css
else
    sass --style compressed style.scss ../style.css
fi
