#!/bin/bash

set -e
set -x

TAG=test

# before_install
#curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
#sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
#sudo apt-get update
#sudo apt-get -y install docker-ce

# install
sudo docker pull rastasheep/ubuntu-sshd
sudo docker pull busybox

docker build -t sshd .travis/ssh
#script

# make the plugin
sudo PLUGIN_TAG=$TAG make
# enable the plugin
sudo docker plugin enable vieux/sshfs:$TAG
# list plugins
sudo docker plugin ls
# start sshd
sudo docker run -d -p 2222:22 sshd

# test1: simple
sudo docker volume create -d vieux/sshfs:$TAG -o sshcmd=root@localhost:/ -o port=2222 -o password=root sshvolume
sudo docker run --rm -v sshvolume:/write busybox sh -c "echo hello > /write/world"
sudo docker run --rm -v sshvolume:/read busybox grep -Fxq hello /read/world
#sudo cat /var/lib/docker/plugins/sshfs-state.json
sudo docker volume rm sshvolume

# test2: allow_other
sudo docker volume create -d vieux/sshfs:$TAG -o sshcmd=root@localhost:/ -o allow_other -o port=2222 -o password=root sshvolume
sudo docker run --rm -v sshvolume:/write -u nobody busybox sh -c "echo hello > /write/world"
docker run --rm -v sshvolume:/read -u nobody busybox grep -Fxq hello /read/world
#sudo cat /var/lib/docker/plugins/sshfs-state.json
sudo docker volume rm sshvolume

# test3: compression
sudo docker volume create -d vieux/sshfs:$TAG -o sshcmd=root@localhost:/ -o Ciphers=arcfour -o Compression=no -o port=2222 -o password=root sshvolume
sudo docker run --rm -v sshvolume:/write busybox sh -c "echo hello > /write/world"
sudo docker run --rm -v sshvolume:/read busybox grep -Fxq hello /read/world
#sudo cat /var/lib/docker/plugins/sshfs-state.json
sudo docker volume rm sshvolume

# test4: source
sudo docker plugin disable vieux/sshfs:$TAG
sudo docker plugin set vieux/sshfs:$TAG state.source=/tmp
sudo docker plugin enable vieux/sshfs:$TAG
sudo docker volume create -d vieux/sshfs:$TAG -o sshcmd=root@localhost:/ -o Ciphers=arcfour -o Compression=no -o port=2222 -o password=root sshvolume
sudo docker run --rm -v sshvolume:/write busybox sh -c "echo hello > /write/world"
sudo docker run --rm -v sshvolume:/read busybox grep -Fxq hello /read/world
#sudo cat /tmp/sshfs-state.json
sudo docker volume rm sshvolume

# test5: ssh key
sudo docker plugin disable vieux/sshfs:$TAG
sudo docker plugin set vieux/sshfs:$TAG sshkey.source=`pwd`/.travis/ssh/
sudo docker plugin enable vieux/sshfs:$TAG
sudo docker volume create -d vieux/sshfs:$TAG -o sshcmd=root@localhost:/ -o port=2222 sshvolume
sudo docker run --rm -v sshvolume:/write busybox sh -c "echo hello > /write/world"
sudo docker run --rm -v sshvolume:/read busybox grep -Fxq hello /read/world
#sudo cat /var/lib/docker/plugins/sshfs-state.json
sudo docker volume rm sshvolume
