FROM node:14-alpine

WORKDIR /workspace
RUN npm install -g image-size@0.7.4 marked@0.6.2

ADD taoblog .
ADD themes themes
ADD setup setup
ADD admin admin


CMD ["./taoblog"]
