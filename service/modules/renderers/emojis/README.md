# 表情包

是的，退后，我要开始~~装逼~~(斗图)了！

## 几个原因

1. 发现微信的表情包在复制的时候会转换成对应的文本描述（比如：`[旺柴]`）
2. 这个写法与 Markdown 标准的参考链接（[Reference Link](https://spec.commonmark.org/0.31.2/#shortcut-reference-link)）写法完全一样
3. 然后 Freya 在[评论](https://blog.twofei.com/622/#comment-1496)中提到了狗头🐶的使用

感觉巧合太多，要不我试一试简单的表情包支持？

## 实现

1. 完全标准的 Markdown 语法，形如：`[旺柴]`，`[皱眉]`
2. 就算不解析成图片，也能见名知义

## 参考

微信表情图片来源：<https://github.com/airinghost/wechat-emoji>。
