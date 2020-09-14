#!/bin/sh



LOCAL_PATH=$(cd `dirname $0`; pwd)
echo "LOCAL_PATH=$LOCAL_PATH"
DECE_PATH="${LOCAL_PATH%}"
echo "DECE_PATH=$DECE_PATH"
CZERO_PATH="${DECE_PATH%/*}/go-czero-import"
echo "CZERO_PATH=$CZERO_PATH"

echo "update go-czero-import"
cd $CZERO_PATH
git fetch&&git rebase

echo "update go-dece"
cd $DECE_PATH
git fetch&&git rebase
make clean
BUILD_PATH="${DECE_PATH%}/build"

os="all"
version="v0.3.1-beta.rc.5"
while getopts ":o:v:" opt
do
    case $opt in
        o)
        os=$OPTARG
        ;;
        v)
        version=$OPTARG
        ;;
        ?)
        echo "unkonw param"
        exit 1;;
    esac
done

if [ "$os" = "all" ]; then
    os_version=("linux-amd64-v3" "linux-amd64-v4" "darwin-amd64" "windows-amd64")
else
    os_version[0]="$os"
fi

for os in ${os_version[@]}
    do
      echo "make gecetx-${os}"
      make "gecetx-"${os}
      rm -rf $BUILD_PATH/gecetxpkg/bin
      rm -rf $BUILD_PATH/gecetxpkg/czero
      mkdir -p $BUILD_PATH/gecetxpkg/bin
      mkdir -p $BUILD_PATH/gecetxpkg/czero/data/
      mkdir -p $BUILD_PATH/gecetxpkg/czero/include/
      mkdir -p $BUILD_PATH/gecetxpkg/czero/lib/
      cp -rf $CZERO_PATH/czero/data/* $DECE_PATH/build/gecetxpkg/czero/data/
      cp -rf $CZERO_PATH/czero/include/* $DECE_PATH/build/gecetxpkg/czero/include/
      if [ $os == "windows-amd64" ];then
        mv $BUILD_PATH/bin/gecetx*.exe $BUILD_PATH/gecetxpkg/bin/tx.exe
        cp -rf  $CZERO_PATH/czero/lib_WINDOWS_AMD64/* $DECE_PATH/build/gecetxpkg/czero/lib/
      elif [ $os == "linux-amd64-v3" ];then
        mv $BUILD_PATH/bin/gecetx-v3* $BUILD_PATH/gecetxpkg/bin/tx
        cp -rf  $CZERO_PATH/czero/lib_LINUX_AMD64_V3/* $DECE_PATH/build/gecetxpkg/czero/lib/
      elif [ $os == "linux-amd64-v4" ];then
        mv $BUILD_PATH/bin/gecetx-v4* $BUILD_PATH/gecetxpkg/bin/tx
        cp -rf  $CZERO_PATH/czero/lib_LINUX_AMD64_V4/* $DECE_PATH/build/gecetxpkg/czero/lib/
      else
        mv $BUILD_PATH/bin/gecetx-darwin* $BUILD_PATH/gecetxpkg/bin/tx
        cp -rf  $CZERO_PATH/czero/lib_DARWIN_AMD64/* $DECE_PATH/build/gecetxpkg/czero/lib/
      fi
      cd $BUILD_PATH

      if [ $os == "windows-amd64" ];then
        rm -rf ./gecetx-*-$os.zip
        zip -r gecetx-$version-$os.zip gecetxpkg/*
      else
         rm -rf ./gecetx-*-$os.tar.gz
         tar czvf gecetx-$version-$os.tar.gz gecetxpkg/*
      fi

      cd $LOCAL_PATH

    done
rm -rf $BUILD_PATH/gecetxpkg/bin
rm -rf $BUILD_PATH/gecetxpkg/czero

