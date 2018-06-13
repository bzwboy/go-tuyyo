#!/bin/sh

. $(pwd)/../conf.sh

ltid="1.1.2.3830.2548.4042"
#ltid="2.1.3033.2725.108.301"
#ltid="err"
#ltid="2.1.1139.742.1597.34"
#ltid="2.1.2.3830.2549.1846"

case "${1}" in
    add)
        echo ">> add cache ..."
        redis-cli -p 6380 --eval lua/add.lua >/dev/null
        myexp
        ;;
    read)
        echo ">> read cache ..."
        redis-cli -p 6380 --eval lua/read.lua
        ;;
    *)
        echo "${1} not support"
esac
