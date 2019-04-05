#!/bin/bash

if ! hash sass 2>/dev/null; then
    echo "sass does not exist.";
    exit 1;
fi

sass --watch sass/prism.scss:prism.css

