all: docker rootfs

docker:
	docker build -t builder -f Dockerfile.dev .
	docker create --name tmp builder
	docker cp tmp:/go/bin/docker-volume-sshfs .
	docker rm -vf tmp
	docker rmi builder
	docker build -t vieux/sshfs:rootfs .

rootfs:
	mkdir -p plugin/rootfs
	docker create --name tmp vieux/sshfs:rootfs
	docker export tmp | tar -x -C ./plugin/rootfs
	sudo cp config.json ./plugin/
	sudo chown -R root ./plugin/
	sudo chgrp -R root ./plugin/
	docker rm -vf tmp


