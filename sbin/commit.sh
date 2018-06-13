#!/bin/sh

. $(pwd)/conf.sh

cd $base_path
echo ">> git add ..."
git add .

echo ">> git status ..."
git st

echo ">> git commit ..."
msg="tmp"
if [ -n "${1}" ]; then
    msg="${1}"
fi 
git ci -m "$msg"

echo ">> git push ..."
git push

echo ">> git status ..."
git st

