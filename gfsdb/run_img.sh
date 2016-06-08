#!/bin/bash
set -e
export PATH=`pwd`:`dirname ${0}`:$PATH
mkdir -p `dirname $5`
mkdir -p `dirname $2`
args=""
if [ "$2" != "" ];then
	args="-resize "$2
fi
convert $1 -resize \>$3x$4 $5
cp -f $5 $2
rm -f $5
echo
echo
echo "----------------result----------------"
echo "[json]"
echo '{"count":1,"files":["'$6'"],"src":"'$1'"}'
echo "[/json]"
echo