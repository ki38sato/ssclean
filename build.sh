#!/bin/bash

VERSION=$(cat ./VERSION)
#PACKAGE="github.com/ki38sato/ssclean"

go get -v github.com/Masterminds/glide
go install github.com/Masterminds/glide
glide up

apt-get update
apt-get install -y zip

GOOS=linux GOARCH=amd64 go build -v -ldflags "-X main.version=${VERSION}" -o build/ssclean
cd build
tar czf ssclean_linux_amd64.tar.gz ssclean
cd ..

GOOS=darwin GOARCH=amd64 go build -v -ldflags "-X main.version=${VERSION}" -o build/ssclean
cd build
zip ssclean_darwin_amd64.zip ssclean
cd ..
