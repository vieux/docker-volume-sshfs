# Docker volume plugin for sshFS

This plugin allows you to mount remote folder using sshfs in your container easily.

[![TravisCI](https://travis-ci.org/vieux/docker-volume-sshfs.svg)](https://travis-ci.org/vieux/docker-volume-sshfs)

## Usage

### Use with SSH keys
1 - Install the plugin

```
$ docker plugin install vieux/sshfs # or docker plugin install vieux/sshfs DEBUG=1
Plugin "vieux/sshfs" is requesting the following privileges:
 - network: [host]
 - mount: [/tmp]
 - device: [/dev/fuse]
 - capabilities: [CAP_SYS_ADMIN]
Do you grant the above permissions? [y/N] y
latest: Pulling from vieux/sshfs
e248a6530152: Download complete
Status: Downloaded newer image for vieux/sshfs:latest
Installed plugin vieux/sshfs
```

2 - Configure driver to point to your SSH keys

_NOTE_: The plugin defaults to looking for SSH keys in `/tmp`. You can copy your `<identity_file_name>` to `/tmp` and omit this step, e.g.
`cp $HOME/.ssh/id_rsa /tmp/.`

```
$ docker plugin disable vieux/sshfs
vieux/sshfs

$ docker plugin set vieux/sshfs KeyPath.source=$HOME/.ssh

$ docker plugin enable vieux/sshfs
vieux/sshfs
```
3 - Create a volume

```
$ docker volume create -d vieux/sshfs -o sshcmd=<user@host:path> -o identity=<identity_file_name> sshvolume
sshvolume

$ docker volume ls
DRIVER              VOLUME NAME
vieux/sshfs         sshvolume
```

4 - Use the volume

```
$ docker run --rm -it -v sshvolume:<path> busybox ls <path>
```

### Alternatively, use with passwords instead of SSH keys
1 - Install the plugin

```
$ docker plugin install vieux/sshfs # or docker plugin install vieux/sshfs DEBUG=1
Plugin "vieux/sshfs" is requesting the following privileges:
 - network: [host]
 - mount: [/tmp]
 - device: [/dev/fuse]
 - capabilities: [CAP_SYS_ADMIN]
Do you grant the above permissions? [y/N] y
latest: Pulling from vieux/sshfs
e248a6530152: Download complete
Status: Downloaded newer image for vieux/sshfs:latest
Installed plugin vieux/sshfs

```

2 - Create a volume

```
$ docker volume create -d vieux/sshfs -o sshcmd=<user@host:path> -o password=<password> sshvolume
sshvolume

$ docker volume ls
DRIVER              VOLUME NAME
vieux/sshfs         sshvolume
```

3 - Use the volume

```
$ docker run --rm -it -v sshvolume:<path> busybox ls <path>
```

### Removing the plugin
_NOTE_: You must remove any volumes created with this plugin prior to removing the plugin itself.

```
$ docker plugin disable vieux/sshfs
vieux/sshfs

$ docker plugin rm vieux/sshfs
vieux/sshfs
```

## Notes
* When using the SSH key approach, the directory where the keys are located must be on the same host as the docker engine.

## THANKS

https://github.com/docker/go-plugins-helpers

## LICENSE

MIT
