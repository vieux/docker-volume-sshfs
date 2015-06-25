# Docker volume plugin for sshFS

This plugin allows you to mount remove folder using sshfs in your container easily.

## Installation

Using go (until we get proper binaries):

```
$ go get github.com/vieux/docker-volume-sshfs
```

## Usage

1 - Start the plugin using this command:

```
$ sudo docker-volume-sshfs
```

2 - Start your docker containers with the option `--volume-driver=sshfs` and use the first part of `--volume` to specify the remote volume that you want to connect to:

```
$ sudo docker run -it --volume-driver sshfs --volume root@1.2.3.4#/data:/data busybox sh
```

Due to a limitation in the docker cli, use `#` instead of `:` to specify the host path.

## THANKS

https://github.com/calavera/dkvolume

## LICENSE

MIT