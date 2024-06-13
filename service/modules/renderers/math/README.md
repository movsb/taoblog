# KaTex

依赖 KaTex 二进制文件，编译方式见 Dockerfile。

`fonts` 是直接从下载的 release 里面拷贝的，只保留 woff2 格式。

为了浏览器不报错，代码内会把 CSS 里面的非 woff2 全部去掉。
