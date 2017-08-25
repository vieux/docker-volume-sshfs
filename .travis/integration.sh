#!/bin/bash

set -e
set -x

TAG=test

# before_install
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
sudo apt-get update
sudo apt-get -y install docker-ce

# install
docker pull rastasheep/ubuntu-sshd:14.04
docker pull busybox

#script

# make the plugin
PLUGIN_TAG=$TAG make
# enable the plugin
docker plugin enable vieux/sshfs:$TAG
# list plugins
docker plugin ls
# start sshd
docker run -d -p 2222:22 rastasheep/ubuntu-sshd:14.04

# test1: simple
docker volume create -d vieux/sshfs:$TAG -o sshcmd=root@localhost:/ -o port=2222 -o password=root sshvolume
docker run --rm -v sshvolume:/write busybox sh -c "echo hello > /write/world"
docker run --rm -v sshvolume:/read busybox grep -Fxq hello /read/world
sudo cat /var/lib/docker/plugins/sshfs-state.json
docker volume rm sshvolume

# test2: allow_other
docker volume create -d vieux/sshfs:$TAG -o sshcmd=root@localhost:/ -o allow_other -o port=2222 -o password=root sshvolume
docker run --rm -v sshvolume:/write -u nobody busybox sh -c "echo hello > /write/world"
docker run --rm -v sshvolume:/read -u nobody busybox grep -Fxq hello /read/world
sudo cat /var/lib/docker/plugins/sshfs-state.json
docker volume rm sshvolume

# test3: compression
docker volume create -d vieux/sshfs:$TAG -o sshcmd=root@localhost:/ -o Ciphers=arcfour -o Compression=no -o port=2222 -o password=root sshvolume
docker run --rm -v sshvolume:/write busybox sh -c "echo hello > /write/world"
docker run --rm -v sshvolume:/read busybox grep -Fxq hello /read/world
sudo cat /var/lib/docker/plugins/sshfs-state.json
docker volume rm sshvolume

# test4: source
docker plugin disable vieux/sshfs:$TAG
docker plugin set vieux/sshfs:$TAG state.source=/tmp
docker plugin enable vieux/sshfs:$TAG
docker volume create -d vieux/sshfs:$TAG -o sshcmd=root@localhost:/ -o Ciphers=arcfour -o Compression=no -o port=2222 -o password=root sshvolume
docker run --rm -v sshvolume:/write busybox sh -c "echo hello > /write/world"
docker run --rm -v sshvolume:/read busybox grep -Fxq hello /read/world
sudo cat /tmp/sshfs-state.json
docker volume rm sshvolume

