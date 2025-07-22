FROM alpine:3.22
RUN apk add ca-certificates
RUN apk add sqlite

# for /etc/mime.types
RUN apk add mailcap

RUN apk add exiftool
RUN apk add libavif-apps 

WORKDIR /workspace
ADD taoblog /usr/local/bin/taoblog
ENTRYPOINT ["taoblog"]
CMD ["server"]
