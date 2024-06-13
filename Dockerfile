FROM alpine:3.20 AS katex
RUN apk add ca-certificates quickjs quickjs-dev gcc musl-dev
RUN wget -O katex.tar.gz https://github.com/KaTeX/KaTeX/releases/download/v0.16.10/katex.tar.gz \
	&& tar xzvf katex.tar.gz
COPY service/modules/renderers/math/katex.js /katex/main.js
RUN cd /katex && qjsc katex.min.js main.js

################################################################################

FROM alpine:3.15
RUN apk add ca-certificates
RUN apk add sqlite
# for /etc/mime.types
RUN apk add mailcap

WORKDIR /workspace
COPY --from=katex /katex/a.out /bin/katex

ADD taoblog taoblog
ENTRYPOINT ["./taoblog"]
CMD ["server"]
