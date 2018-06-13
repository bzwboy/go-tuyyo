#!/bin/sh

. $(pwd)/conf.sh

cd $base_path

echo ">> create wechat_prod ..."
cp bin/wechat bin/wechat_prod
myexp