#!/bin/sh

. $(pwd)/../conf.sh

echo ">> start wechat service ..."

if [ -f /tmp/wechat_dev.pid ]; then
    ps -ef |grep -w $(cat /tmp/wechat_dev.pid) |grep -v grep >/dev/null
    if [ $? -eq 0 ]; then
        echo "+Ok, process has existed!"
        exit $?
    fi
fi

../../bin/wechat -m dev -n longtooth -d /tmp -f /home/ubuntu/tuyyo/etc/conf_dev.ini >>/tmp/debug 2>&1 &
myexp
