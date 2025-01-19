#!/bin/sh
DEST="$1"
SRC=$(realpath -m "$DEST" | sed -e "s:^/usr/share:/share:")
ln -svf "$PREFIX/$SRC" "$DEST"
