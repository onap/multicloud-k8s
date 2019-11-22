#!/bin/bash

# usage: build_images.sh

set -ex
base_image_tag=openwrt-1806-4-base
docker_file=Dockerfile_1806_mwan3
image_tag=openwrt-1806-mwan3
package=openwrt-18.06.4-x86-64-generic-rootfs

# build openwrt base docker images
base_image=`docker images | grep $base_image_tag | awk '{print $1}'`
if [ -z "$base_image" ]; then
    # download driver source package
    if [ ! -e /tmp/$package.tar.gz ]; then
        wget -P /tmp https://downloads.openwrt.org/releases/18.06.4/targets/x86/64/$package.tar.gz
    fi
    cp /tmp/$package.tar.gz .

    docker import $package.tar.gz $base_image_tag
fi

# generate Dockerfile
test -f ./set_proxy && . set_proxy
docker_proxy=${docker_proxy-""}
if [ -z "$docker_proxy" ]; then
    cp ${docker_file}_noproxy.tpl $docker_file
else
    cp $docker_file.tpl $docker_file
    sed -i "s,{docker_proxy},$docker_proxy,g" $docker_file
fi

# build docker images for openwrt with wman3
docker build --network=host -f $docker_file -t $image_tag .

# clear
docker image rm $base_image_tag
rm -rf $docker_file
rm -rf $package.tar.gz
