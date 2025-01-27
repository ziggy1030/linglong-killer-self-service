#!/bin/bash
set -e
ROOT_DIR=$(dirname $(readlink -f $0))
rm -rf ldd-found.log
ll-builder build | grep --line-buffered -Pv "={3,}"

echo Runing ldd-check...
$ROOT_DIR/ll-killer ldd-check | tee ldd-check.log
CHECK=$(cat ldd-check.log || true)
if [ -z "$CHECK" ]; then
    exit 0
fi
echo Runing ldd-search...
$ROOT_DIR/ll-killer ldd-search ldd-check.log ldd-found.log ldd-missing.log

FOUND=$(cat ldd-found.log || true)
if [ -n "$FOUND" ]; then
    echo Found libraries:
    cat ldd-found.log
    ll-builder build | grep --line-buffered -Pv "={3,}"
    echo Recheck ldd...
    $ROOT_DIR/ll-killer ldd-check | tee ldd-check.log
fi
