# TODO

## 关于 base URL

文章的链接形如：`/123/`，页面的链接形如：`/about`，或者 `/about/`。
内容中的相对链接对于文章来说完全没问题，比如 `main.go` 最终为 `/123/main.go`。

但是页面就不行了：

- `/about` → `/main.go`
- `/about/` → `/about/main.go`

只有后者可以正确解析。
但是如果页面和文件都有目录层级，比如页面是 `/about/sub/`，文件是 `dir/main.go`。
最终的链接是 `/about/sub/dir/main.go`。
后端服务无法/难以识别出哪一部分属于页面路径、哪一部分属于文件路径。

所以，目前所有的文章/页面下都包含了一个 `<base href="/123/">` 类似这样的标签，
用于告知浏览器页面中所有的相对链接的 `baseURI`。
但是这会[破坏锚点链接（`#section`）的点击](https://developer.mozilla.org/en-US/docs/Web/HTML/Element/base#in-page_anchors)：

> Links pointing to a fragment in the document — e.g. `<a href="#some-id">` — are resolved with the `<base>`, triggering an HTTP request to the base URL with the fragment attached.
>
> For example, given `<base href="https://example.com/">` and this link: `<a href="#anchor">To anchor</a>`. The link points to <https://example.com/#anchor>.

点击锚点时，浏览器会修改地址栏为 `<base>` + 锚点，体验并不好。

还有一些其它解决办法：

- 渲染 Markdown 成 HTML 的时候将页面的锚点绝对地址。

  比如，页面地址是 `/about`，base 是 `/123/`，锚点是 `#name`。
  那么，修改前，锚点的绝对地址是 `/123/#name`，这会导致浏览器重新加载。
  而如果渲染时把锚点改成 `/about#name`，就不会重新加载了。
