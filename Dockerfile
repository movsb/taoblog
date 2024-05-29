FROM alpine:3.15

RUN apk add ca-certificates
RUN apk add sqlite
# for /etc/mime.types
RUN apk add mailcap

WORKDIR /workspace

ADD taoblog taoblog
ENTRYPOINT ["./taoblog"]
CMD ["server"]
