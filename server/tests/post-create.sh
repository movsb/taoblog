#!/bin/bash

cat << EOF | http -j POST :2564/v1/posts
{
    "title": "Title",
    "source": "Source",
    "source_type": "html",
    "tags": [
        "tag1",
        "tag2"
    ]
}
EOF
