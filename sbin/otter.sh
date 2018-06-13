#!/bin/sh

. $(pwd)/conf.sh

help () {
    echo "$0 [start|stop|status|restart|reload|build|install]"
    exit 127
}

if [ -z "${1}" ]; then
    help
fi

install() {
    echo ">> Install Gin Package ..."
    go get -v -u github.com/gin-gonic/gin
    myexp

    echo ">> Install Errors Package ..."
    go get -v -u github.com/pkg/errors
    myexp

    echo ">> Install Redis Package ..."
    go get -v -u github.com/gomodule/redigo/redis
    myexp
}

start() {
    echo ">> Start LtCenter Service ..."
    if [ ! -f /tmp/wechat_product.pid ]; then
        touch /tmp/wechat_product.pid
    fi

    ps -ef |grep $(cat /tmp/wechat_product.pid) |grep -v grep >/dev/null
    if [ $? -eq 0 ]; then
        echo "process exist"
        exit $?
    fi

    if [ ! -f ../bin/wechat_prod ]; then
        sh build_prod.sh >/dev/null
    fi

    nohup ../bin/wechat_prod -m product &
    myexp
}

stop() {
    echo ">> Stop wechat Service ..."
    kill -QUIT $(cat /tmp/wechat_product.pid)
    myexp
}

restart() {
    stop
    sleep 1
    start
}

reload() {
    echo ">> Reload wechat service ..."
    kill -USR1 $(cat /tmp/wechat_product.pid)
    myexp
}

status() {
    echo ">> Longtooth status ..."
    kill -USR2 $(cat /tmp/wechat_product.pid)
    myexp
}

case "${1}" in
    install)
        install
        ;;

    start)
        start
        ;;

    stop)
        stop
        ;;

    restart)
        restart
        ;;

    status)
        status
        ;;

    reload)
        reload
        ;;

    build)
        sh ./build.sh
        ;;

    *)
        help
        ;;
esac
