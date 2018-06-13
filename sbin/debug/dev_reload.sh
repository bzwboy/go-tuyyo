#!/bin/bash

. $(pwd)/../conf.sh

echo ">> reload process ..."
kill -USR1 $(cat /tmp/wechat_dev.pid)
myexp

