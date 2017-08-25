#!/bin/bash

set -e
set -x

# install
docker pull rastasheep/ubuntu-sshd:14.04
docker pull busybox

#script

# make the plugin
make
# enable the plugin
make enable
# list plugins
docker plugin ls
# start sshd
docker run -d -p 2222:22 rastasheep/ubuntu-sshd:14.04

# test1: simple
docker volume create -d vieux/sshfs:next -o sshcmd=root@localhost:/ -o port=2222 -o password=root sshvolume
docker run --rm -v sshvolume:/write busybox sh -c "echo hello > /write/world"
docker run --rm -v sshvolume:/read busybox grep -Fxq hello /read/world
sudo cat /var/lib/docker/plugins/sshfs-state.json
docker volume rm sshvolume

# test2: allow_other
docker volume create -d vieux/sshfs:next -o sshcmd=root@localhost:/ -o allow_other -o port=2222 -o password=root sshvolume
docker run --rm -v sshvolume:/write -u nobody busybox sh -c "echo hello > /write/world"
docker run --rm -v sshvolume:/read -u nobody busybox grep -Fxq hello /read/world
sudo cat /var/lib/docker/plugins/sshfs-state.json
docker volume rm sshvolume

# test3: compression
docker volume create -d vieux/sshfs:next -o sshcmd=root@localhost:/ -o Ciphers=arcfour -o Compression=no -o port=2222 -o password=root sshvolume
docker run --rm -v sshvolume:/write busybox sh -c "echo hello > /write/world"
docker run --rm -v sshvolume:/read busybox grep -Fxq hello /read/world
sudo cat /var/lib/docker/plugins/sshfs-state.json
docker volume rm sshvolume

# test4: source
docker plugin disable vieux/sshfs:next
docker plugin set vieux/sshfs:next state.source=/tmp
docker plugin enable vieux/sshfs:next
docker volume create -d vieux/sshfs:next -o sshcmd=root@localhost:/ -o Ciphers=arcfour -o Compression=no -o port=2222 -o password=root sshvolume
docker run --rm -v sshvolume:/write busybox sh -c "echo hello > /write/world"
docker run --rm -v sshvolume:/read busybox grep -Fxq hello /read/world
sudo cat /tmp/sshfs-state.json
docker volume rm sshvolume

