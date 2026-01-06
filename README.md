# taoblog

我为自己编写的博客系统。

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
$ make tools protos generate test build'
```

### 日常开发

直接`go build`就行，或者直接运行：`go run main.go server`。

开箱即用，不需要任何配置文件。

## 目录结构

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
