#!/bin/sh

tarfile="gopub-$1-linux-amd64.tar.gz"

echo "开始打包$tarfile..."

export GOARCH=amd64
export GOOS=linux

bee pack -exs=".go:.DS_Store:.tmp:.log" -exr=data

mv gopub.tar.gz $tarfile
