#!/bin/bash -e
PROC="demo"

rm -rf ewhine_pkg
mkdir -p ewhine_pkg

#CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' .

git log|head -n 1 > version
go build -o $PROC

mv "$PROC" ewhine_pkg/
cp control.sh ewhine_pkg/mx_$PROC
cp $PROC.conf start-stop-daemon version build.rb ewhine_pkg/

tar zcvf $PROC.tar.gz ewhine_pkg
