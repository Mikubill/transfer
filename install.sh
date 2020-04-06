#!/usr/bin/env bash

set -e

hash tar uname grep curl head
OS="$(uname)"
case $OS in
  Linux)
    OS='linux'
    ;;
  Darwin)
    OS='darwin'
    ;;
  *)
    echo 'OS not supported'
    exit 2
    ;;
esac

ARCH="$(uname -m)"
case $ARCH in
  x86_64|amd64)
    ARCH='amd64'
    ;;
  aarch64)
    ARCH='arm64'
    ;;
  i?86|x86)
    ARCH='386'
    ;;
  arm*)
    ARCH='arm'
    ;;
  *)
    echo 'OS type not supported'
    exit 2
    ;;
esac

DOWNLOAD_URL=$(curl -fsSL https://api.github.com/repos/Mikubill/transfer/releases/latest | grep "browser_download_url.*$OS.*$ARCH" | cut -d '"' -f 4)

curl -L "$DOWNLOAD_URL" | tar xz

printf "\nTransfer Downloded.\n\n"
exit 0