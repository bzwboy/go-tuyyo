#!/bin/sh

port="8081"
if [ -n "${1}" ]; then
    port="${1}"
fi

echo ">>Ping localhost:$port ..."
ret=$(curl --connect-timeout 1 http://localhost:${port}/ping 2>/dev/null)
echo $ret