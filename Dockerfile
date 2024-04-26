FROM alpine:3.15

RUN apk add ca-certificates
RUN apk add sqlite

WORKDIR /workspace

ADD admin/login.html admin/
ADD admin/profile.html admin/
ADD admin/script.js  admin/
ADD theme/blog/statics theme/blog/statics/
ADD theme/blog/templates theme/blog/templates/
ADD taoblog taoblog

ENTRYPOINT ["./taoblog"]
CMD ["server"]
