FROM alpine

RUN apk update && apk add sshfs

RUN mkdir -p /run/docker/plugins

COPY docker-volume-sshfs docker-volume-sshfs

CMD ["docker-volume-sshfs"]