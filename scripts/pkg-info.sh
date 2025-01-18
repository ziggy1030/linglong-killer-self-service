#!/bin/bash
field="$1"
cat package.info | sed -n "/^$field:/,/^[^ ]/p" | sed -E -e "s/^$field://" | grep '^\s' | sed -e 's/^\s*//'
