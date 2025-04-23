# 同步/备份数据到 Cloudflare R2 免费存储

## 需要同步的数据

* posts.db

  直接定时原样存储。

* files.db

  文件太大，不一次性上传。

  鉴于 R2 的 A/B 类操作免费次数有限，这里采取以下方案：

  * 每篇文章只打包存储为一个文件；
  * 附件有修改时，整体打包重新上传；

## 存储规划

* 免费空间：10GB
* 付费空间：$1.35/100GB/月

感觉非常便宜，可以按日期定期存储。
[低频存储][lf]的价格只比标准版便宜一丁点，而且还自在 Beta 测试中。暂时可以不考虑使用。

配置好“Object lifecycle rules”和“Bucket lock rules”以自动删除过期对象和防止对象被误删，备份程序不会主动删除文件。

[lf]: https://developers.cloudflare.com/r2/pricing/#r2-pricing

价格计算器：<https://r2-calculator.cloudflare.com/>

## 注意

* 插件第一次运行时会全量上传，后续只会增量上传
* 不要手动修改 R2 仓库内的文件结构

## 安全

所有文件默认加密保存。

## 已知的问题

- [ ] 已经删除的文章附件不会被删除
