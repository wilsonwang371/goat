FROM ubuntu:latest
ARG TARGETARCH

RUN apt-get update

WORKDIR /root
COPY goat-$TARGETARCH /usr/local/bin/goat
COPY samples /etc/goat/samples
