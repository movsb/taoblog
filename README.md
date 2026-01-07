# taoblog

我为自己编写的博客系统。

## 功能与特点

* 零依赖、零配置、跨平台、单二进制、高覆盖率的端到端测试；
* 嵌入式SQLite数据库(无CGO)，所有资源文件也存数据库，自动升级数据库；
* 统一接口：一份API，同时支持前端(主题)、本地客户端、手机客户端(没续费已下架)调用；
* 一键本地增量备份、定时整站备份、自动同步资源到对象存储（私有文章资源加密后存储）；
* 用户名密码登录（支持一次性验证）、通行密钥登录（人脸识别、指纹识别）；
* 多用户系统、文章权限支持：公开、私有、站内分享、草稿；
* 自带搜索引擎：可检测文章标题、文章内容、图片内容（考虑支持中）；
* 严格兼容CommonMark、GFM，图表(PlantUML、Tldraw、DrawIO、ECharts、GraphViz)，数学公式；
* 微信九宫格照片、照片Exif信息展示、实况照片、自动模糊预加载、自动边框；
* 支持日历：纪念日、提醒事项、行程记录、课程表、公历/农历生日提醒、文章周期提醒回顾；
* 自带评论：发表、回复、编辑、删除、转移(到其它文章）、邮件通知、手机即时通知；
* 图片直接本地上传、自动高质量压缩，自动同步到第三方图床，自动按国家取外链；
* 自动 git 同步：文章内容修改会自动增量提交到 git 仓库；
* 自动检测域名过期状态、自动检测证书过期状态（临期提醒）；
* 完全的主题能力开放、自定义主题编写功能；
* 自动生成 OpenGraph、X/Twitter/推特分享图片；
* 一键导入推特数据，其它平台后期考虑支持；
* ……

## 关于项目

主要目的是为了学习前端与后端知识，我非常喜欢折腾，目的算是达到了。另外一个原因是：我之前用过 WordPress，但我发现它的功能太过复杂完善，快要不适合我了。

我不是前端开发者，这个项目主要是我零碎时间学习前、后端知识的时候顺带写成。
博客程序后端最开始完全由 PHP 写成，后使用 Go 语言完全重构。（PHP 依旧是世界上最好的语言，是我不够好。）
前端由纯JavaScript + Go语言模板写成，与 React/Vue 等框架无关。（主要原因是作为长期项目，稳定为主，跟不上它们的节奏。）

很感谢你们对此项目的 watch，star 与 fork。
如果你们跟我一样不是前端开发者，但同样对前端有比较浓厚的兴趣，我希望你们加我为好友，一起来学习。
如果你也有兴趣来编写一套你自己的（或是对大家的）博客系统，或任何与之相关的话题，非常高兴能与你讨论。

## 如果你想试一下

可以一句话就启动起来我的博客系统（需要 Docker 哦！）：

```bash
$ docker run -it --rm --name=taoblog -p 2564:2564 taocker/taoblog:amd64-latest
```

然后打开：<http://localhost:2564>。

## 本地开发

依赖的工具：

|工具|描述|地址|
|---|---|---|
|`make`|用于执行 make 命令||
|`go`|Go语言编译工具链|[All releases - The Go Programming Language](https://go.dev/dl/)。<br>安装完后记得把`$HOME/go/bin`添加到`$PATH`。|
|`protoc`|用来编译Protocol Buffers。|[Protocol Buffer Compiler Installation \| Protocol Buffers Documentation](https://protobuf.dev/installation/)|
|`bun`|Bun JavaScript运行时，用来打包。|[Bun — A fast all-in-one JavaScript runtime](https://bun.sh/)|
|`sass`|用来预编译 SCSS 样式文件。|[Sass: Install Sass](https://sass-lang.com/install/)|

不依赖 cgo。

### 克隆项目到本地

```bash
$ git clone https://github.com/movsb/taoblog
```

### 首次编译

建议完整跑一遍：

```bash
$ make tools protos generate test build
```

### 日常开发

直接`go build`就行，或者直接运行：`go run main.go server`。

开箱即用，不需要任何配置文件。

## 目录结构

代码目录结构：

|文件名|文件描述|
|------|--------|
|admin/      | 后台目录|
|cmd/        | 客户端/服务端命令/配置|
|docker/     | 容器镜像|
|gateway/    | 网关接口层|
|modules/    | 公共模块|
|protocols/  | 协议定义|
|service/    | 服务实现|
|setup/      | 安装管理|
|theme/      | 主题目录|
|tests/      | 端到端测试|
|main.go     | 入口程序|

运行时目录结构：

```bash
taoblog $ tree 
.
├── access.log
├── cache.db
├── files.db
└── posts.db

1 directory, 4 files
```
