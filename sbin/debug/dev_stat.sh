#!/bin/sh

. $(pwd)/../conf.sh

kill -USR2 $(cat /tmp/wechat_dev.pid)
