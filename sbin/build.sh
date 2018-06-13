#!/bin/sh

. $(pwd)/conf.sh

cd $base_path

echo ">> update git ..."
git pull -f
myexp

echo ">> build wechat project ..."
#go build -o ../bin/wechat wechat
go install -work wechat
myexp

#echo ">> build httpd project ..."
##go build -o ../bin/httpd httpd
#go install -work httpd
#myexp
