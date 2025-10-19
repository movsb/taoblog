FROM ubuntu:24.04

RUN apt update -y && apt install -y curl unzip make
RUN apt install -y ca-certificates
RUN apt install -y git
RUN git config --global --add safe.directory /workspace

ADD install-protoc.sh .
RUN ./install-protoc.sh

ADD install-sass.sh .
RUN ./install-sass.sh

ADD install-go.sh .
RUN ./install-go.sh
ENV PATH="$PATH":/usr/local/go/bin:/root/go/bin

# 安装 Swift 以编译 swift。
# RUN curl -O https://download.swift.org/swift-5.10-release/ubuntu2204/swift-5.10-RELEASE/swift-5.10-RELEASE-ubuntu22.04.tar.gz \
# 	&& tar xzvf swift-5.10-RELEASE-ubuntu22.04.tar.gz \
# 	&& ln -s "$(pwd)"/swift-5.10-RELEASE-ubuntu22.04/usr/bin/swift /usr/bin/swift \
# 	&& rm swift-5.10-RELEASE-ubuntu22.04.tar.gz
# RUN echo '/swift-5.10-RELEASE-ubuntu22.04/usr/lib/swift/linux' > /etc/ld.so.conf.d/swift.conf \
# 	&& ldconfig
# RUN apt install -y sqlite3 libncurses5-dev
# RUN curl -LO https://github.com/grpc/grpc-swift/releases/download/1.23.0/protoc-grpc-swift-plugins-linux-x86_64-1.23.0.zip \
# 	&& unzip protoc-grpc-swift-plugins-linux-x86_64-1.23.0.zip \
# 	&& mv protoc-gen-grpc-swift protoc-gen-swift /usr/bin \
# 	&& rm protoc-grpc-swift-plugins-linux-x86_64-1.23.0.zip

# 安装 Node.js v23（通过 apt）
# RUN curl -fsSL https://deb.nodesource.com/setup_23.x | bash - && apt install -y nodejs

# 安装 BUN
RUN curl -fsSL https://bun.sh/install | bash
ENV PATH="/root/.bun/bin:$PATH"

# 真是奇怪，代码写了很久了，为什么安装了 bun 以后就不见了？
# service/modules/renderers/image/image_test.go
RUN apt install media-types
