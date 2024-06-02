# media_size

媒体文件（图片、视频）大小/尺寸计算器。

## 目的

给 `<img>` 这种加上 `width` 和 `height` 特性值，防止图片在加载成功前未知大小而显示为零宽大小。
这样会导致图片加载成功 page flow 发生变化，感观效果不好。也不方便前端通过 CSS 控制过宽、过高的图片的显示方式。

## SVG

SVG 的 ViewBox 只是描述原始默认的大小尺寸，其会根据 ViewPort（Container）自动缩放。
所以用 ViewBox 作为大小其实是不合适的，还应该加上 scale 参数。
