#!/bin/bash

ROOT_DIR=$(dirname $(readlink -f $0))
APP_SOCK=app.sock
ROOT_FS_DIR=/rootfs
UNIX_SOCKET=/$APP_SOCK
LL_KILLER_EXEC=$ROOT_DIR/ll-killer
LL_KILLER_TIMEOUT=30s
LL_KILLER_ARGS=" --unix $UNIX_SOCKET --unix-timeout $LL_KILLER_TIMEOUT"

mkdir -p $ROOT_FS_DIR
$LL_KILLER_EXEC $LL_KILLER_ARGS --stack "$ROOT_FS_DIR:/:$ROOT_DIR" --stack $ROOT_FS_DIR/usr/share:/usr/share:$ROOT_DIR/share --chroot "$ROOT_FS_DIR" -- "${@:-bash}"
