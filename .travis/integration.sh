#!/bin/bash

set -e
set -x

# install
docker pull rastasheep/ubuntu-sshd:14.04
docker pull busybox

#script
make
make enable
docker plugin ls
docker run -d -p 2222:22 rastasheep/ubuntu-sshd:14.04
docker volume create -d vieux/sshfs:next -o sshcmd=root@localhost:/ -o port=2222 -o password=root sshvolume
docker volume ls
docker run -it -v sshvolume:/data1 busybox sh -c "echo hello > /data1/world"
docker run -it -v sshvolume:/data2 busybox grep -Fxq hello /data2/world
