FROM ubuntu:latest

RUN apt-get update

WORKDIR /root
COPY goat-$TARGETARCH /usr/local/bin/goat
COPY samples /etc/goat/samples
