# 版本发布系统

当前版本：v2.0.1

基于Git的代码发布系统，用于发布PHP等脚本语言开发的项目。使用Go语言和Beego框架开发。本人所在公司已使用了半年，累计超过五百次发版，到目前为止没出过什么问题，现在功能已经比较完善。

## 功能

1. 多帐号、多角色、权限管理
2. 发版邮件通知、邮件模板设置
3. 支持多个项目，每个项目可设置多个发布环境
4. 支持发版前、发版后执行指定shell脚本
5. 支持自动生成版本号文件
6. 支持发版审批，可针对不同项目选择开启

## 流程

整个发版流程如下：

1. 发布系统构建发布包
2. 将发布包发布到跳板机
3. 在跳板机进行解压，将代码同步到目标服务器。

## 下载地址

- [https://github.com/lisijie/gopub/releases](https://github.com/lisijie/gopub/releases)

## 安装

仅支持linux/mac系统，并且要求安装了mysql和git。

安装步骤：

1. 创建数据库，将install.sql导入mysql。
2. 修改 conf/app.conf 中相关的配置。
3. 使用命令 `./service.sh start` 启动，如果无法启动，检查主程序 gopub 是否具有可执行权限，使用 `chmod +x ./gopub` 增加权限。
4. 使用 `http://localhost:8000` 访问。
5. 后台默认帐号为 `admin`，密码为 `admin888`。 

## 使用docker运行

在源码目录使用docker-compose启动即可。

	$ docker-compose up

## 界面截图

![gopub](https://raw.githubusercontent.com/lisijie/gopub/master/screenshot.png)
