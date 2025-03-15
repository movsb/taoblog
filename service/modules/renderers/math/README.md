# KaTex

`fonts` 是直接从下载的 release 里面拷贝的，只保留 woff2 格式。
为了浏览器不报错，代码内会把 CSS 里面的非 woff2 全部去掉。

* `katex.min.{css,js}` 和 `fonts` 来源于：<https://github.com/KaTeX/KaTeX/releases/tag/v0.16.21>。

[Javy: JavaScript → WASM](https://github.com/bytecodealliance/javy)

* `quickjs.wasm` 生成：`javy emit-plugin > quickjs.wasm`。
* `katex.bundle.js` 生成：`esbuild --bundle --minify --outfile=katex.bundle.js katex.js`。需安装：`go install github.com/evanw/esbuild/cmd/esbuild@latest`。
* `katex.wasm` 生成：`javy build -C dynamic=y -C plugin=binary/quickjs.wasm -C  source-compression=y -o binary/katex.wasm katex.bundle.js`。

TODO: 写脚本自动生成上述内容。
