FROM alpine:3.8

RUN apk add ca-certificates

WORKDIR /workspace

ADD taoblog .
ADD admin admin
ADD themes themes
ADD service service
ADD setup setup

CMD ["./taoblog"]
