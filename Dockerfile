FROM alpine:3.20 AS sass
RUN wget -O sass.tgz 'https://github.com/sass/dart-sass/releases/download/1.82.0/dart-sass-1.82.0-linux-x64.tar.gz'
RUN mkdir -p /tmp/sass && tar xzvf sass.tgz -C /tmp/sass

################################################################################

FROM alpine:3.15
RUN apk add ca-certificates
RUN apk add sqlite
# for /etc/mime.types
RUN apk add mailcap
RUN apk add exiftool

WORKDIR /workspace
COPY --from=sass /tmp/sass/dart-sass /opt/sass
RUN cd /usr/local/bin && ln -s /opt/sass/sass

ADD taoblog /usr/local/bin/taoblog
ENTRYPOINT ["taoblog"]
CMD ["server"]
