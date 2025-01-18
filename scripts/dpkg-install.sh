#!/bin/bash
set -e

if [ ! -z "$1" ]; then
    echo "last process exited with code $?, fallback to dpkg-install..."
fi
source $(dirname $0)/env.sh
dpkg -i $LL_SOURCES_DIR/*.deb
