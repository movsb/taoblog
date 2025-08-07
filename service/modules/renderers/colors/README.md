# Colors

## 方案

```html
<color></color>

<!-- by class -->
<color red></color>
<color :blue></color>
<color red:blue></color>

<!-- by css color name and function -->
<color fg="red" bg="blue"></color>
<color fg="rgb(1,1,1)" bg="hsl(330,100%,50%)"></color>
```

## 颜色来源(colors.yaml)

* 来源：[140 html color names as an array of json objects—enjoy!](https://gist.github.com/jennyknuth/e2d9ee930303d5a5fe8862c6e31819c5)
* 经 `cat colors.json | yq -P` 转换成 yaml
  * 去掉不常用的
  * 去掉不能同时在黑底、白底下良好显示的
  * 调整了部分颜色（比如蓝色、黄色，以不至于太复古）

## 参考

* [CSS Named Colors: Groups, Palettes, Facts, & Fun](https://austingil.com/css-named-colors/)
* [CSS Color Module Level 4](https://www.w3.org/TR/css-color-4/#named-colors)
* [Web Content Accessibility Guidelines (WCAG) 2.1](https://www.w3.org/TR/WCAG21/#contrast-minimum)
* [Color contrast - Accessibility | MDN](https://developer.mozilla.org/en-US/docs/Web/Accessibility/Guides/Understanding_WCAG/Perceivable/Color_contrast?utm_source=devtools&utm_medium=a11y-panel-checks-color-contrast)
* [140 html color names as an array of json objects—enjoy!](https://gist.github.com/jennyknuth/e2d9ee930303d5a5fe8862c6e31819c5)
