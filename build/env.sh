#!/bin/sh
set -e

if [ ! -f "build/env.sh" ]; then
    echo "$0 must be run from the root of the repository."
    exit 2
fi


root="$PWD"

# Create fake Go workspace if it doesn't exist yet.
workspace="$PWD/build/_workspace"

godir="$workspace/src/github.com/dece-cash/go-dece"

CZERO_PATH="$godir/czero"

#rm -rf "$workspace"

go mod vendor


mkdir -p "$godir"
##
cp -R ` find .   \( -path "./build" -o -path "./.git" -o -path "./tests" -o -path "./Makefile" -o -path "./.idea" -o -path "./go.sum" -o -path "./README.md"  -o -path "./go.mod" -o -path "./makepkg.sh" -o -path "./maketxpkg.sh"  \) -prune -o -type d -depth 1 -print ` "$godir"

cp -R interfaces.go "$godir"

GOPATH="$workspace"

export GOPATH

cd "$root"
args=()
index=0
for i in "$@"; do
   args[$index]=$i
   index=$[$index+1]
done


function sysname() {

    SYSTEM=`uname -s |cut -f1 -d_`

    if [ "Darwin" == "$SYSTEM" ]
    then
        echo "Darwin"

    elif [ "Linux" == "$SYSTEM" ]
    then
        kernal=`uname -v |cut -f 1 -d \ |cut -f 2 -d -`
        if [ "Ubuntu"  == "$kernal" ]
        then
            echo "Linux-V3"
        else

            name=`uname  -r |cut -f1 -d.`
            echo Linux-V"$name"
        fi
    else
        echo "$SYSTEM"
    fi



}

SNAME=`sysname`

if [ "Darwin" == "$SNAME" ]
then
    DYLD_LIBRARY_PATH="$CZERO_PATH/lib/lib_DARWIN_AMD64"
    export DYLD_LIBRARY_PATH
elif [ "Linux-V3" == "$SNAME" ]
then
   LD_LIBRARY_PATH="$CZERO_PATH/lib/lib_LINUX_AMD64_V3"
   export LD_LIBRARY_PATH
elif [ "Linux-V4" == "$SNAME" ]
then
    LD_LIBRARY_PATH="$CZERO_PATH/lib/lib_LINUX_AMD64_V4"
    export LD_LIBRARY_PATH
elif [ "$SNAME" == "Linux-*" ]
then
     echo "only support linux kernal v3 or v4"
     exit
elif [ "$SNAME" == "MINGW32" ]
then
    cd "$CZERO_PATH"
    cp -rf lib_WINDOWS_AMD64/* lib/
else
   echo "only support Mingw"
   exit
fi

cd "$root"

if [ $1 == "linux-v3" ]; then
    cd "$CZERO_PATH"
    cp -rf lib_LINUX_AMD64_V3/* lib/
    cd "$root"
    unset args[0]
elif [ $1 == "linux-v4" ];then
    cd "$CZERO_PATH"
    cp -rf lib_LINUX_AMD64_V4/* lib/
    cd "$root"
    unset args[0]
elif [ $1 == "darwin-amd64" ];then
     cd "$CZERO_PATH"
     cp -rf lib_DARWIN_AMD64/* lib/
     cd "$root"
     unset args[0]
elif [ $1 == "windows-amd64" ];then
    unset args[0]
    cd "$CZERO_PATH"
    cp -rf lib_WINDOWS_AMD64/* lib/
    cd "$root"
else
     echo "local"
fi


# Run the command inside the workspace.
cd "$DECE_PATH"
PWD="$DECE_PATH"

#Launch the arguments with the configured environment.
exec "${args[@]}"