# LivePhoto - 实况照片

* 调研：<https://blog.twofei.com/1603/>
* 文档：<https://developer.mozilla.org/en-US/docs/Web/Media/Guides/Autoplay>

使用：完全无感。正常使用图片，如果有同名但后缀为 `.mp4` 的文件，自动渲染成 LivePhoto。

## 注意

因为 Live Photo 在显示的时候一般来说尺寸大小相对固定，所以不宜将其放在九宫格图中，会挤压导致形变，也不利于预览。

因此，优先使用以下代码：

```markdown
![](image.jpg)
```

而不是：

```markdown
<gallery>

![](image.jpg)
![](image.avif)

</gallery>
```
