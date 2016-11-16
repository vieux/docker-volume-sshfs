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
	docker rm -vf tmp

all: docker rootfs
