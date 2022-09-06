FROM ubuntu:latest

RUN apt-get update

WORKDIR /root
COPY goat /usr/local/bin/goat
