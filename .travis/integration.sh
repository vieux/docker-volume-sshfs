#!/bin/bash

set -e
set -x

TAG=test

docker build -t sshd .travis/ssh

# make the plugin
PLUGIN_TAG=$TAG make
# enable the plugin
PLUGIN_TAG=$TAG make enable
# list plugins
docker plugin ls
# start sshd
docker run -d -p 2222:22 sshd

echo "# test1: simple"
docker volume create -d vieux/sshfs:$TAG -o sshcmd=root@localhost:/ -o port=2222 -o password=root sshvolume
docker run --rm -v sshvolume:/write busybox sh -c "echo hello > /write/world"
docker run --rm -v sshvolume:/read busybox grep -Fxq hello /read/world
#cat /var/lib/docker/plugins/sshfs-state.json
docker volume rm sshvolume

echo "# test2: allow_other"
docker volume create -d vieux/sshfs:$TAG -o sshcmd=root@localhost:/ -o allow_other -o port=2222 -o password=root sshvolume
docker run --rm -v sshvolume:/write -u nobody busybox sh -c "echo hello > /write/world"
docker run --rm -v sshvolume:/read -u nobody busybox grep -Fxq hello /read/world
#cat /var/lib/docker/plugins/sshfs-state.json
docker volume rm sshvolume

echo "# test3: compression"
docker volume create -d vieux/sshfs:$TAG -o sshcmd=root@localhost:/ -o Ciphers=arcfour -o Compression=no -o port=2222 -o password=root sshvolume
docker run --rm -v sshvolume:/write busybox sh -c "echo hello > /write/world"
docker run --rm -v sshvolume:/read busybox grep -Fxq hello /read/world
#cat /var/lib/docker/plugins/sshfs-state.json
docker volume rm sshvolume

echo "# test4: source"
docker plugin disable vieux/sshfs:$TAG
docker plugin set vieux/sshfs:$TAG state.source=/tmp
docker plugin enable vieux/sshfs:$TAG
docker volume create -d vieux/sshfs:$TAG -o sshcmd=root@localhost:/ -o Ciphers=arcfour -o Compression=no -o port=2222 -o password=root sshvolume
docker run --rm -v sshvolume:/write busybox sh -c "echo hello > /write/world"
docker run --rm -v sshvolume:/read busybox grep -Fxq hello /read/world
#cat /tmp/sshfs-state.json
docker volume rm sshvolume

echo "# test5: ssh key"
docker plugin disable vieux/sshfs:$TAG
docker plugin set vieux/sshfs:$TAG sshkey.source=`pwd`/.travis/ssh/
docker plugin enable vieux/sshfs:$TAG
docker volume create -d vieux/sshfs:$TAG -o sshcmd=root@localhost:/ -o port=2222 sshvolume
docker run --rm -v sshvolume:/write busybox sh -c "echo hello > /write/world"
docker run --rm -v sshvolume:/read busybox grep -Fxq hello /read/world
#cat /var/lib/docker/plugins/sshfs-state.json
docker volume rm sshvolume
