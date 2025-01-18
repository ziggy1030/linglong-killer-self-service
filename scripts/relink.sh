#!/bin/sh
DEST="$1"
ls -l $DEST
SRC=$(realpath -m "$DEST" | sed -e "s:^/usr/share:/share:")
ln -svf "$PREFIX/$SRC" "$DEST"
