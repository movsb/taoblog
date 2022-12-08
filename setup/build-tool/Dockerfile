FROM ubuntu:22.10

RUN apt update -y && apt install -y curl unzip
RUN apt install -y make gcc
RUN apt install -y ca-certificates

ADD install-go.sh .
RUN ./install-go.sh
ENV PATH="$PATH":/usr/local/go/bin:/root/go/bin

ADD install-protoc.sh .
RUN ./install-protoc.sh

ADD install-sass.sh .
RUN ./install-sass.sh

RUN apt install -y git
RUN git config --global --add safe.directory /workspace
