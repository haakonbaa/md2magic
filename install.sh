#!/bin/sh
set -o errexit

go build .
ln -sf $PWD/md2magic /usr/bin/md2magic
