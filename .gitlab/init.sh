#!/bin/bash
# Initializes project directory for build / test / etc...
go get -u github.com/golang/dep/cmd/dep
mkdir -p $GOPATH/src
cd $GOPATH/src
ln -s $CI_PROJECT_DIR
cd $CI_PROJECT_NAME
dep ensure