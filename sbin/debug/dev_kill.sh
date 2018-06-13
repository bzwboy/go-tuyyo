#!/bin/bash

. $(pwd)/../conf.sh

echo ">> kill process ..."
if [ ! -f /tmp/wechat_dev.pid ]; then
    echo "+Ok, Succ.\n"
    exit 0
fi

#kill -QUIT $(cat /tmp/wechat_dev.pid)
ps -ef |grep -w $(cat /tmp/wechat_dev.pid) |grep -v grep >/dev/null
if [ $? -ne 0 ]; then
    echo "+Ok, process not exist.\n"
    exit 0
fi

kill $(cat /tmp/wechat_dev.pid)
myexp

