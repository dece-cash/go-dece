#! /bin/sh

_GOPATH=`cd ../../../../../;pwd`

export GOPATH=$_GOPATH
echo $GOPATH

go install -v ../cmd/gece
go install -v ../cmd/decekey
