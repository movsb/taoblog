FROM alpine:3.8

RUN apk add ca-certificates
RUN apk add mysql-client
RUN apk add sqlite

WORKDIR /workspace

ADD server-linux-amd64 taoblog
ADD admin admin
ADD themes themes
ADD service service
ADD setup setup

CMD ["./taoblog"]
