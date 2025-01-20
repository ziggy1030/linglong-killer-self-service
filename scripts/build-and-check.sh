#!/bin/bash
set -e
ROOT_DIR=$(dirname $(readlink -f $0))
ll-builder build | grep --line-buffered -Pv "={3,}"
echo Runing ldd-check...
ll-builder run --exec "entrypoint.sh $ROOT_DIR/ldd-check.sh" | tr -d '\r' | tee ldd-check.log

$ROOT_DIR/ll-killer ldd-check >ldd-check.log
$ROOT_DIR/ll-killer ldd-search ldd-check.log ldd-found.log ldd-missing.log

FOUND=$(cat ldd-found.log || true)
if [ -n "$FOUND" ]; then
    echo Found libraries:
    cat ldd-found.log
    ll-builder build | grep --line-buffered -Pv "={3,}"
    $ROOT_DIR/ll-killer ldd-check >ldd-check.log
fi
