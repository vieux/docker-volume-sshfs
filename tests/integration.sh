#!/bin/bash

set -e
set -x

TAG=test


# install
docker pull rastasheep/ubuntu-sshd
docker pull busybox

docker build -t sshd tests/testdata
#script

# make the plugin
PLUGIN_TAG=$TAG make
# enable the plugin
docker plugin enable vieux/sshfs:$TAG
# list plugins
docker plugin ls
# start sshd
docker run --name sshd -d -p 2222:22 sshd

# test1: simple
docker volume create -d vieux/sshfs:$TAG -o sshcmd=root@localhost:/ -o port=2222 -o password=root sshvolume
docker run --rm -v sshvolume:/write busybox sh -c "echo hello > /write/world"
docker run --rm -v sshvolume:/read busybox grep -Fxq hello /read/world
docker volume rm sshvolume

# test2: allow_other
docker volume create -d vieux/sshfs:$TAG -o sshcmd=root@localhost:/ -o allow_other -o port=2222 -o password=root sshvolume
docker run --rm -v sshvolume:/write -u nobody busybox sh -c "echo hello > /write/world"
docker run --rm -v sshvolume:/read -u nobody busybox grep -Fxq hello /read/world
docker volume rm sshvolume

# test3: compression
docker volume create -d vieux/sshfs:$TAG -o sshcmd=root@localhost:/ -o Ciphers=arcfour -o Compression=no -o port=2222 -o password=root sshvolume
docker run --rm -v sshvolume:/write busybox sh -c "echo hello > /write/world"
docker run --rm -v sshvolume:/read busybox grep -Fxq hello /read/world
docker volume rm sshvolume

# test4: restart
docker volume create -d vieux/sshfs:$TAG -o sshcmd=root@localhost:/ -o port=2222 -o password=root sshvolume
docker run --rm -v sshvolume:/write busybox sh -c "echo hello > /write/world"
(sleep 2; docker restart sshd) &
docker run --rm -v sshvolume:/read busybox sh -c "sleep 4 ; grep -Fxq hello /read/world"
docker volume rm sshvolume

# test5: source
docker plugin disable vieux/sshfs:$TAG
docker plugin set vieux/sshfs:$TAG state.source=/tmp
docker plugin enable vieux/sshfs:$TAG
docker volume create -d vieux/sshfs:$TAG -o sshcmd=root@localhost:/ -o Ciphers=arcfour -o Compression=no -o port=2222 -o password=root sshvolume
docker run --rm -v sshvolume:/write busybox sh -c "echo hello > /write/world"
docker run --rm -v sshvolume:/read busybox grep -Fxq hello /read/world
docker volume rm sshvolume

# test6: ssh key
docker plugin disable vieux/sshfs:$TAG
docker plugin set vieux/sshfs:$TAG sshkey.source=`pwd`/tests/testdata/
docker plugin enable vieux/sshfs:$TAG
docker volume create -d vieux/sshfs:$TAG -o sshcmd=root@localhost:/ -o port=2222 sshvolume
docker run --rm -v sshvolume:/write busybox sh -c "echo hello > /write/world"
docker run --rm -v sshvolume:/read busybox grep -Fxq hello /read/world
docker volume rm sshvolume
