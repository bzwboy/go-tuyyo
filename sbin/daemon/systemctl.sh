#!/bin/sh

. $(pwd)/../conf.sh

# install
install () {
    echo ">> Install systemctl service ..."
	sudo cp ../../etc/systemctl/wechat.service /lib/systemd/system/wechat.service
	sudo systemctl enable wechat
	myexp
}

# service
stop() {
    echo ">> Stop wechat service ..."
    sudo systemctl stop wechat
	myexp
}

start() {
    echo ">> Start wechat service ..."
    sudo systemctl start wechat
	myexp
}

restart() {
    echo ">> Restart wechat service ..."
    sudo systemctl restart wechat
	myexp
}

$1
