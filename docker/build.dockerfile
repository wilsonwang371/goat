FROM golang:1.18
ARG TARGETARCH

RUN apt-get update
RUN go install mvdan.cc/gofumpt@latest
RUN go install golang.org/x/tools/cmd/goimports@latest
RUN apt update && apt install -qy npm
RUN npm install -g prettier@latest

WORKDIR /root
COPY go.mod .
COPY go.sum .
RUN go mod download
RUN rm /root/go.mod
RUN rm /root/go.sum
