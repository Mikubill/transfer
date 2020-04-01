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

VERSION=$(curl https://api.github.com/repos/Mikubill/transfer/releases/latest 2>&1 | grep -o "[0-9]\.[0-9]\.[0-9]" | head -n 1)

curl -L https://github.com/Mikubill/transfer/releases/download/v$VERSION/transfer_$VERSION\_$OS\_$ARCH.tar.gz | tar xz

printf "\nTrabsfer Downloded.\n\n"
exit 0