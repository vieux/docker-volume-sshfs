FROM alpine

RUN apk update && apk add sshfs

RUN mkdir -p /run/docker/plugins /mnt/state /mnt/volumes

COPY docker-volume-sshfs docker-volume-sshfs

CMD ["docker-volume-sshfs"]
