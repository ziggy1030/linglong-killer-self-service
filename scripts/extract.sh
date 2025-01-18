#!/bin/bash
set -e
if [ ! -z "$1" ]; then
    echo "last process exited with code $?, fallback to extract..."
fi
source $(dirname $0)/env.sh
for deb in $LL_SOURCES_DIR/*.deb; do
    dpkg -x $deb $PREFIX
done
