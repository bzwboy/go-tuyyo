#!/bin/sh

. $(pwd)/../conf.sh

if [ "$1" != "restart" ]; then
    cd $(pwd)/../
    ./build.sh
    sleep 1

    cd $(pwd)/debug
fi

sh dev_kill.sh

sleep 1
sh dev_start.sh

