#!/bin/sh



LOCAL_PATH=$(cd `dirname $0`; pwd)
echo "LOCAL_PATH=$LOCAL_PATH"
DECE_PATH="${LOCAL_PATH%}"
echo "DECE_PATH=$DECE_PATH"

echo "update go-dece"

cd $DECE_PATH
git fetch&&git rebase


BUILD_PATH="${DECE_PATH%}/build"

CZERO_PATH="${DECE_PATH%}/czero"

os="all"
version=`date "+%Y-%m-%d"`
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
    os_version=("linux-amd64" "darwin-amd64" "windows-amd64")
else
    os_version[0]="$os"
fi

for os in ${os_version[@]}
    do
      echo "make gece-${os}"
      make "gece-"${os}
      rm -rf $BUILD_PATH/gecepkg/bin
      rm -rf $BUILD_PATH/gecepkg/czero
      mkdir -p $BUILD_PATH/gecepkg/bin
      mkdir -p $BUILD_PATH/gecepkg/czero/data/
      mkdir -p $BUILD_PATH/gecepkg/czero/include/
      mkdir -p $BUILD_PATH/gecepkg/czero/lib/
      cp -rf $CZERO_PATH/lib/data/* $DECE_PATH/build/gecepkg/czero/data/
#      cp -rf $CZERO_PATH/czero/include/* $DECE_PATH/build/gecepkg/czero/include/
      if [ $os == "windows-amd64" ];then
        mv $BUILD_PATH/bin/gece*.exe $BUILD_PATH/gecepkg/bin/gece.exe
        cp -rf  $CZERO_PATH/lib/lib_WINDOWS_AMD64/* $DECE_PATH/build/gecepkg/czero/lib/
      elif [ $os == "linux-amd64" ];then
#        mv $BUILD_PATH/bin/bootnode-v3*  $BUILD_PATH/gecepkg/bin/bootnode
        mv $BUILD_PATH/bin/gece-linux* $BUILD_PATH/gecepkg/bin/gece
        cp -rf  $CZERO_PATH/lib/lib_LINUX_AMD64_V3/* $DECE_PATH/build/gecepkg/czero/lib/
      else
        mv $BUILD_PATH/bin/gece-darwin* $BUILD_PATH/gecepkg/bin/gece
        cp -rf  $CZERO_PATH/lib/lib_DARWIN_AMD64/* $DECE_PATH/build/gecepkg/czero/lib/
      fi
      cd $BUILD_PATH

      if [ $os == "windows-amd64" ];then
        rm -rf ./gece-*-$os.zip
        zip -r gece-$version-$os.zip gecepkg/*
      else
         rm -rf ./gece-*-$os.tar.gz
         tar czvf gece-$version-$os.tar.gz gecepkg/*
      fi

      cd $LOCAL_PATH

    done
rm -rf $BUILD_PATH/gecepkg/bin
rm -rf $BUILD_PATH/gecepkg/czero

