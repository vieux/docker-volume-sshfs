#!/bin/sh

remote_name=vieux/sshfs

set -e

id=$(docker create "$remote_name" true)
mkdir -p /var/lib/docker/plugins/$id/rootfs
docker export "$id" | tar -x -C /var/lib/docker/plugins/$id/rootfs
docker rm -vf "$id"


#Create a sample manifest file
cat <<EOF > /var/lib/docker/plugins/$id/manifest.json
{
	"manifestVersion": "v0.1",
	"description": "sshFS plugin for Docker",
	"documentation": "https://docs.docker.com/engine/extend/plugins/",
	"entrypoint": ["/go/bin/docker-volume-sshfs"],
	"interface" : {
		"types": ["docker.volumedriver/1.0"],
		"socket": "sshfs.sock"
	}
	"capabilities": ["CAP_SYS_ADMIN"]
}
EOF
