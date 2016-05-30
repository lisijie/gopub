# 版本发布系统

当前版本：v2.0.0

基于Git的代码发布系统，用于发布PHP等脚本语言开发的项目。使用Go语言和Beego框架开发。本人所在公司已使用了半年，累计超过五百次发版，到目前位置没出过什么问题，现在功能已经比较完善。

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

- v2.0.0 Linux版： [http://pan.baidu.com/s/1jIMwYSA](http://pan.baidu.com/s/1jIMwYSA "gopub-linux-v2.0.0.zip")

## 安装

仅支持linux/mac系统，并且要求安装了mysql和git。

安装步骤：

1. 创建数据库，将install.sql导入mysql
2. 使用命令 `./service.sh start` 启动
3. 使用 `http://localhost:8000` 访问

## 界面截图

![gopub](https://raw.githubusercontent.com/lisijie/gopub/master/screenshot.png)
