FROM alpine:3.15

RUN apk add ca-certificates
RUN apk add sqlite

WORKDIR /workspace

ADD admin admin
ADD setup setup
ADD themes themes
ADD protocols protocols
ADD taoblog taoblog

ENTRYPOINT ["./taoblog"]
CMD ["server"]
