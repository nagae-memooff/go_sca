#!/bin/bash -e


NAME="$1"

if [ "$NAME" == "" ]; then
  echo "usage: $0 proname"
  exit 1
fi

git clone git@git.sabachat.cn:golang/go_sca.git $NAME
cd $NAME

sed -i "s/PROG=demo/PROG=$NAME/g" demo.conf
mv demo.conf $NAME.conf

sed -i "s/Proname = \"demo\"/Proname = \"$NAME\"/g" info.go
