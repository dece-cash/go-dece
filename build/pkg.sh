#!/bin/sh

LOCAL_PATH=$(cd `dirname $0`; pwd)
DECE_PATH="${LOCAL_PATH%/*}"
CZERO_PATH="${DECE_PATH%/*}/go-czero-import"

echo "update go-czero-import"
cd $CZERO_PATH
git fetch&&git rebase

echo "update go-dece"
cd $DECE_PATH
git fetch&&git rebase
make clean all

rm -rf $LOCAL_PATH/gecepkg/bin
rm -rf $LOCAL_PATH/gecepkg/czero
mkdir -p $LOCAL_PATH/gecepkg/czero/data/
mkdir -p $LOCAL_PATH/gecepkg/czero/include/
mkdir -p $LOCAL_PATH/gecepkg/czero/lib/
cp -rf $LOCAL_PATH/bin $LOCAL_PATH/gecepkg
cp -rf $CZERO_PATH/czero/data/* $DECE_PATH/build/gecepkg/czero/data/
cp -rf $CZERO_PATH/czero/include/* $DECE_PATH/build/gecepkg/czero/include/

function sysname() {

    SYSTEM=`uname -s |cut -f1 -d_`

    if [ "Darwin" == "$SYSTEM" ]
    then
        echo "Darwin"

    elif [ "Linux" == "$SYSTEM" ]
    then
        name=`uname  -r |cut -f1 -d.`
        echo Linux-V"$name"
    else
        echo "$SYSTEM"
    fi
}

SNAME=`sysname`

if [ "Darwin" == "$SNAME" ]
then
    echo $SNAME
    cp $CZERO_PATH/czero/lib_DARWIN_AMD64/* $DECE_PATH/build/gecepkg/czero/lib/
elif [ "Linux-V3" == "$SNAME" ]
then
    echo $SNAME
    cp $CZERO_PATH/czero/lib_LINUX_AMD64_V3/* $DECE_PATH/build/gecepkg/czero/lib/
elif [ "Linux-V4" == "$SNAME" ]
then
    echo $SNAME
    cp $CZERO_PATH/czero/lib_LINUX_AMD64_V4/* $DECE_PATH/build/gecepkg/czero/lib/
fi

cd $LOCAL_PATH
if [ -f ./gecepkg_*.tar.gz ]; then
	rm ./gecepkg_*.tar.gz
fi
tar czvf gecepkg_$SNAME.tar.gz gecepkg/*
