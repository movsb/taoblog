FROM node:14-alpine

WORKDIR /workspace
RUN npm install -g image-size@0.7.4 marked@0.6.2

ADD taoblog .
ADD themes themes
ADD setup setup
ADD admin admin
ADD service service

ENV NODE_PATH=/usr/local/lib/node_modules

CMD ["./taoblog"]
