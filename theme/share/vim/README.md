# vim

[KeyboardEvent: altKey property - Web APIs | MDN](https://developer.mozilla.org/en-US/docs/Web/API/KeyboardEvent/altKey)

* 在MacOS上，`altKey`是`Option`（不常用），`metaKey`是`Cmd`（更常用）。
* 在Windows上，`altKey`是`Alt`（更常用），`metaKey`是`Windows Logo`（不常用）。

自动化处理：`altKey`和`metaKey`只同时使用一个。在MacOS上处理为`Cmd`，在Windows上处理为`Alt`。

缩写规则：

* `c` 表示 control
* `a` 表示 alt/cmd
* `m` 不使用
* `s` 不使用，区分大小写可以直接用`a`或`A`表示。
