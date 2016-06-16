#!/usr/bin/env python3

# ~tao @ 2016-06-17 04:52
# usage: ./convert.py < input.json > output.json

import sys
import json

obj = json.load(fp=sys.stdin)

content = ''
for stdout in obj['stdout']:
    content += stdout[1]


ret = {}
ret['width'] = obj['width']
ret['height'] = obj['height']
ret['stdout'] = content

print(json.dumps(ret))

