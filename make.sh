#!/bin/bash

rm -rf vendor/*
##https://github.com/Masterminds/glide/issues/654

echo "execute build.sh using golang:1.10"
docker run --rm -v "$(pwd)":/go/src/github.com/ki38sato/ssclean -w /go/src/github.com/ki38sato/ssclean golang:1.10 bash build.sh
