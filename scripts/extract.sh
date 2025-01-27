#!/bin/bash
set -e
if [ ! -z "$1" ]; then
    echo "[fallback] last process exited with code $1, fallback to extract..." >&2
fi
source $(dirname $0)/env.sh
for deb in $LL_SOURCES_DIR/*.deb; do
    dpkg -x $deb $PREFIX
done
