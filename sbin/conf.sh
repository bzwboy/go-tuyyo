#!/bin/sh

# Setting
base_path="$HOME/tuyyo"

export LD_LIBRARY_PATH=$base_path/src/clib
export GOPATH=$base_path
export GOBIN=$base_path/bin
export GIN_MODE=debug #release

# func
myexp() {
    if [ $? -ne 0 ]; then
        echo "-Err, happen wrong\n"
        exit $?
    else
        echo "+Ok, succ.\n"
    fi
}

