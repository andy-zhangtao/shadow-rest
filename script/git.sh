#!/bin/bash
rm -f ../shadowsocks/handler/versionHandler.go

git rev-list HEAD | sort > config.git-hash

LOCALVER=`wc -l config.git-hash | awk '{print $1}'`
if [ $LOCALVER \> 1 ] ; then
    VER=`git rev-list origin/master | sort | join config.git-hash - | wc -l | awk '{print $1}'`
    if [ $VER != $LOCALVER ] ; then
        VER="$VER+$(($LOCALVER-$VER))"
    fi
    if git status | grep -q "modified:" ; then
        VER="${VER}M"
    fi
    VER="$VER $(git rev-list HEAD -n 1 | cut -c 1-7)"
    GIT_VERSION=r$VER
else
    GIT_VERSION=
    VER="x"
fi

rm -f config.git-hash
 
cat ../shadowsocks/handler/versionHandler.go.template | sed "s/\$GIT/$GIT_VERSION/g" > ../shadowsocks/handler/versionHandler.go
 
echo "Generated version"