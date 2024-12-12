FROM alpine:3.20 AS katex
RUN apk add ca-certificates quickjs quickjs-dev gcc musl-dev
RUN wget -O katex.tar.gz https://github.com/KaTeX/KaTeX/releases/download/v0.16.10/katex.tar.gz \
	&& tar xzvf katex.tar.gz
COPY service/modules/renderers/math/katex.js /katex/main.js
RUN cd /katex && qjsc katex.min.js main.js

################################################################################

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
COPY --from=katex /katex/a.out /usr/local/bin/katex
COPY --from=sass /tmp/sass/dart-sass /opt/sass
RUN cd /usr/local/bin && ln -s /opt/sass/sass

ADD taoblog /usr/local/bin/taoblog
ENTRYPOINT ["taoblog"]
CMD ["server"]
