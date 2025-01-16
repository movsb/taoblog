# 测试页面

## 二号标题

### 三号标题

#### 四号标题

##### 五号标题

###### 六号标题

---

## 文本

### 段落

这是段落。

这是第二段。

### 引用

普通文本引用内容。

> A block quotation (also known as a long quotation or extract) is a quotation in a written document, that is set off from the main text as a paragraph, or block of text.
>
> It is typically distinguished visually using indentation and a different typeface or smaller size quotation. It may or may not include a citation, usually placed at the bottom.

### 列表

有序列表：

1. 第一条
2. 第二条
   1. 第二、一条
   2. 第二、二条
3. 第三条

无序列表：

* 第一条
* 第二条
  * 第二、一条
  * 第二、二条
* 第三条

### 详细/展开

* 鼠标放在“总结”上面应该有箭头指示展开方向。

<details>
<summary>总结</summary>

详细内容。

更详细的内容。

</details>

### 水平分隔线

---

### 表格

* 表头应该是 sticky 的。

| Tables   |      Are      |  Cool |
|----------|:-------------:|------:|
| col 1 is |  left-aligned | $1600 |
| col 2 is |    centered   |   $12 |
| col 3 is | right-aligned |    $1 |

### 代码

* 键盘：<kbd>Ctrl</kbd> + <kbd>C</kbd>
* 行内代码：`<div>code</div>`
* 块级代码：

  ```go
  // You can edit this code!
  // Click here and start typing.
  package main
  
  import "fmt"
  
  func main() {
  	fmt.Println("Hello, 世界")
  }
  ```

### 行内元素

[这是一个链接](https://example.com)。
**这是粗体文本**，*这是斜体文本*，~~这是删除线文本~~。

## 嵌入元素

### 图片

不存在的图片：

![](https://not-found/a.jpg)

不存在、带 alt 的图片：

![不存在的图片文本](https://not-found/a.jpg)

## 表单

### 输入框

* 文本：<input type=text />
* 密码：<input type=password />
* 网址：<input type=url />
* 邮箱：<input type=email />
* 日期：<input type=datetime-local />
* 编辑：<textarea></textarea>

### 按钮

* 提交：<input type=submit value="提交按钮" />
* 重置：<input type=reset value="重置按钮" />
* 按钮：<input type=button value="输入按钮" />
* 按钮：<button>普通按钮</button>

### 选择

下拉框：<select>
	<option>选项一</option>
	<option>选项二</option>
	<option>选项三</option>
</select>

复选框：<label><input type=checkbox />选项一</label>
<label><input type=checkbox />选项二</label>
<label><input type=checkbox />选项三</label>

单选框：<label><input type=radio name="r" />选项一</label>
<label><input type=radio name="r" />选项一</label>
<label><input type=radio name="r" />选项一</label>

## 其它

### 对话框

<button onclick="document.querySelector('#dialog').showModal()">显示</button>

<dialog id="dialog">
  <p>一段文字。</p>
  <form method="dialog">
    <input type="submit" value="关闭" />
  </form>
</dialog>
