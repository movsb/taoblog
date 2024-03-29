FROM alpine:3.15

RUN apk add ca-certificates
RUN apk add sqlite

WORKDIR /workspace

ADD admin/login.html admin/
ADD setup/data setup/data/
ADD theme/blog/statics theme/blog/statics/
ADD theme/blog/templates theme/blog/templates/
ADD protocols/docs protocols/docs/
ADD taoblog taoblog

ENTRYPOINT ["./taoblog"]
CMD ["server"]
