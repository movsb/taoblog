# KaTex

版本：<https://github.com/KaTeX/KaTeX/releases/tag/v0.16.21>。

`fonts` 是直接从下载的 release 里面拷贝的，只保留 woff2 格式。

* 为了浏览器不报错，代码内会把 CSS 里面的非 woff2 全部去掉（正则：`,url[^}]+`）。
* 然后把 fonts 路径修改为正确的路径：`fonts/` → `/v3/dynamic/katex/fonts/`。
* `gzip -9c katex.min.js > katex.min.js.gz`。
