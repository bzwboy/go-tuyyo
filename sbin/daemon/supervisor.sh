#!/bin/sh

. $(pwd)/../conf.sh

# install
install () {
    echo ">> Install supervisor package ..."
	apt install supervisor
	myexp
}

# start supervisord
start_supervisor() {
    echo ">> Start supervisord service ..."
    supervisord -c /etc/supervisor/supervisord.conf
	myexp
}

# service
stop() {
    echo ">> Stop wechat service ..."
    supervisorctl stop wechat:wechat_prod
	myexp
}

start() {
    echo ">> Start wechat service ..."
    supervisorctl start wechat:wechat_prod
	myexp
}

restart() {
    echo ">> Restart wechat service ..."
    supervisorctl restart wechat:wechat_prod
	myexp
}

$1
