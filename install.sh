#!/bin/sh
set -o errexit

go build .
cp ./md2magic /usr/bin/
