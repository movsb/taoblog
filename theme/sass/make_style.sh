#!/bin/bash

if ! hash sass 2>/dev/null; then
    echo "sass does not exist.";
    exit 1;
fi

sass --style compressed style.scss ../style.css

