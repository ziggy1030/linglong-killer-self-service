#!/bin/bash
FOUND=${2:-/dev/stdout}
MISSING=${3:-/dev/stderr}
INPUT=${1:--}
SEARCHED=$(cat $INPUT | xargs -P$(nproc) -I{} sh -c 'apt-file find -x "{}$"| grep -P "^lib|/usr/lib/x86_64-linux-gnu/" | sort -urk2 | head -n1')
grep -Fvf <(echo "$SEARCHED" | awk -F/ '{print $NF}') ldd-check.log >$MISSING
echo "$SEARCHED" | cut -d: -f1 | sort -u >$FOUND
