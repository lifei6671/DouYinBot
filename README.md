# DouYinBot

抖音(中国区)无水印视频、背景音乐、作者ID、作者昵称、作品标题等的全能解析和下载。

## 写在前面

- 本项目纯属个人爱好创作
- 所有视频的版权始终属于「字节跳动」
- 严禁用于任何商业用途，如果构成侵权概不负责

## 目前功能

- 解析无水印视频
- 解析视频标题
- 解析作者昵称
- 解析作者ID
- 不需要去除多余字符
- 微信公众号消息转发后解析
- 解析视频存库(仅支持通过微信转发消息的抖音视频，仅支持sqlite数据库)
- 视频上传到七牛
- 视频首页列表展示

## 使用

### 编译

```shell
go build -o douyinbot main.go
```

### 运行

```shell
./douyinbot --config-file=配置文件 --data-file=数据库路径
```


### Docker 使用

#### 部署 [ChromeDouYin](https://github.com/lifei6671/ChromeDouYin) 项目

```go
go install github.com/lifei6671/ChromeDouYin

```

默认情况下 [ChromeDouYin](https://github.com/lifei6671/ChromeDouYin) 会自动下载一个无头浏览器，并通过无头浏览器抓取抖音信息。

但是不保证所有系统都能成功，因此建议使用Docker部署：

```shell
docker run -p 7317:7317 ghcr.io/go-rod/rod
```

部署成功后，  [ChromeDouYin](https://github.com/lifei6671/ChromeDouYin) 会自动连接到该实例。

#### 部署 DouYinBot

```shell
docker pull lifei6671/douyinbot:v1.0.17
docker run -p 9080:9080 -v /data/conf:/var/www/douyinbot/conf /data/data:/var/www/douyinbot/data -v /data/douyin:/var/www/douyinbot/douyin -d lifei6671/douyinbot:v1.0.18
```

需要修改配置文件中的代理信息：

```
douyinproxy=ChromeDouYin的访问接口，如果配置了认证信息只支持https访问
douyinproxyusername=认证用户名
douyinproxypassword=认证密码
```