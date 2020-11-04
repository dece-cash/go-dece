#!/bin/sh
set -e

if [ ! -f "build/env.sh" ]; then
    echo "$0 must be run from the root of the repository."
    exit 2
fi


root="$PWD"

# Create fake Go workspace if it doesn't exist yet.
workspace="$PWD/build/_workspace"

godir="$workspace/src/github.com/dece-cash"

CZERO_PATH="$godir/go-dece/czero"

#rm -rf "$workspace"

go mod vendor


root="$PWD"
if [ ! -L "$godir/go-dece" ]; then
    mkdir -p "$godir"
    cd "$godir"
    ln -s ../../../../../. go-dece
    cd "$root"
fi


##
#cp -R ` find .   \( -path "./build" -o -path "./.git" -o -path "./tests" -o -path "./Makefile" -o -path "./.idea" -o -path "./go.sum" -o -path "./README.md"  -o -path "./go.mod" -o -path "./makepkg.sh" -o -path "./maketxpkg.sh"  \) -prune -o -type d -depth 1 -print ` "$godir"

#cp -R interfaces.go "$godir"
#
GOPATH="$workspace"

export GOPATH

cd "$root"
args=()
index=0
for i in "$@"; do
   args[$index]=$i
   index=$[$index+1]
done


rm -rf $CZERO_PATH/lib/lib/*
cd "$CZERO_PATH/lib/"
cp -rf lib_DARWIN_AMD64/* lib/
cp -rf lib_LINUX_AMD64_V3/* lib/
cp -rf lib_WINDOWS_AMD64/* lib/

DYLD_LIBRARY_PATH="$CZERO_PATH/lib/lib_DARWIN_AMD64"
export DYLD_LIBRARY_PATH

LD_LIBRARY_PATH="$CZERO_PATH/lib/lib_LINUX_AMD64_V3"
export LD_LIBRARY_PATH


# Run the command inside the workspace.
cd "$godir/go-dece"
PWD="$godir/go-dece"

#Launch the arguments with the configured environment.
exec "${args[@]}"