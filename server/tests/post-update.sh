#!/bin/bash

cat << EOF | http -j POST :2564/v1/posts/1
{
    "id": 737,
    "title": "Title",
    "source": "Source",
    "source_type": "html",
    "tags": [
        "tag1",
        "tag2"
    ]
}
EOF
