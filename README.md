# taoblog

我为自己编写的博客系统。

## 编写目的

主要目的是为了学习前端与后端知识，我非常喜欢折腾，目的算是达到了。

另外一个原因是：我之前用过 WordPress，但我发现它的功能太过复杂完善，快要不适合我了。

## 关于项目

我不是前端开发者，这个项目主要是我零碎时间学习前后端知识的时候顺带写成。我完全没有打算把她做成一个大家都能使用的博客系统，相反，她带有浓重的个人风格。
换句话说，除了我以外，可能没有其他人能够很好地使用这套系统。但我尽可能地使得她更加易于使用。

博客程序最开始完全由 PHP 写成，后使用 Go 语言完全重构。（PHP 依旧是世界上最好的语言，是我不够好。）

## 致对项目感兴趣的你们

很感谢你们对此项目的 watch，star 与 fork。

如果你们跟我一样不是前端开发者，但同样对前端有比较浓厚的兴趣，我希望你们加我为好友，一起来学习。

如果你也有兴趣来编写一套你自己的（或是对大家的）博客系统，或任何与之相关的话题，非常高兴能与你讨论。

## 文件说明

文件名|文件描述
------|--------
admin/      | 后台目录（仅登录）
client/     | 博客客户端
config/     | 配置模块
docker/     | 容器镜像
gateway/    | 网关接口层
modules/    | 公共模块
protocols/  | 协议定义
server/     | 博客后台
service/    | 服务实现
setup/      | 安装管理
run/        | 临时目录
themes/     | 主题目录
main.go     | 入口程序

## RoadMap

- 抄几个主题：

  - <http://www.templex.xyz/> 同王垠的博客
  - <https://www.tbfeng.com/> 天边风
  - <https://www.v2ex.com/t/561257> 你见过最简约美观的技术博客

## 联系我

- QQ: 191035066
- EM: chkesp@gmail.com

## 如果你想试一下

可以一句话就启动起来我的博客系统（需要 Docker 哦！）：

```bash
$ docker run -it --rm --name=taoblog -p 2564:2564 -p 2563:2563 taocker/taoblog:amd64-latest
```

然后打开：<http://localhost:2564>。

## 安装

### 依赖

#### protoc

<https://github.com/protocolbuffers/protobuf/releases/latest>

#### sass

<https://github.com/sass/dart-sass/releases/latest>
