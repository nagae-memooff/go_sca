#!/bin/bash -e


NAME="$1"

if [ "$NAME" == "" ]; then
  echo "usage: $0 proname"
  exit 1
fi

git clone git@github.com:nagae-memooff/go_sca.git $NAME
cd $NAME

sed -i "s/PROG=demo/PROG=$NAME/g" control.sh
mv demo.conf $NAME.conf

sed -i "s/Proname = \"demo\"/Proname = \"$NAME\"/g" info.go

sed -i "s/PROC=\"demo\"/PROC=\"$NAME\"/g" build.sh
sed -i "s/progname = \"demo\"/progname = \"$NAME\"/g" build.rb


sed -i "s/go_sca/$NAME/g" api/api.go
sed -i "s/go_sca/$NAME/g" biz/user.go
sed -i "s/go_sca/$NAME/g" entities/user.go
sed -i "s/go_sca/$NAME/g" main.go
sed -i "s/go_sca/$NAME/g" info.go
sed -i "s/go_sca/$NAME/g" go.mod

rm -rf .git
rm go_sca.sh
