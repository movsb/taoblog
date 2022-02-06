FROM ubuntu:22.04

RUN apt update -y && apt install -y curl unzip
RUN apt install -y make gcc git

ADD install-go.sh .
RUN ./install-go.sh

ADD install-protoc.sh .
RUN ./install-protoc.sh

ADD install-sass.sh .
RUN ./install-sass.sh
