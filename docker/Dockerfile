#FROM alpine:3.8
FROM frolvlad/alpine-glibc:alpine-3.8_glibc-2.28

RUN apk add ca-certificates
RUN apk add mysql-client
RUN apk add sqlite

WORKDIR /workspace

ADD taoblog taoblog
ADD admin admin
ADD themes themes
ADD setup setup

CMD ["./taoblog"]
