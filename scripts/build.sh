#!/bin/bash
set -e
function pkg_info_local() {
    field="$1"
    cat package.info | sed -n "/^$field:/,/^[^ ]/p" | sed -E -e "s/^$field://" | grep '^\s' | sed -e 's/^\s*//'
}
PKG=$(pkg_info_local Package)
PKG_VERSION=$(pkg_info_local Version)
if [ -n "$PKG_VERSION" ]; then
    PKG="$PKG=$PKG_VERSION"
fi
IFS=$' ,\n' read -r -a PKGS <<<"$(pkg_info_local Depends)"
apt update -y
dpkg --configure -a
apt install --no-upgrade -yf "$PKG" "${PKGS[@]}"
