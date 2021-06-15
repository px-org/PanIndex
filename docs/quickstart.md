# 快速开始
- [在线演示](https://t1.netrss.cf "https://t1.netrss.cf")

## 平台支持
由于PanIndex交叉编译需要cgo（sqlite），目前很多平台还不能很好的支持，如果你有特殊的编译需求，请告知我，我会尽量添加
- Linux （x86 / amd64 / arm / arm64 ）
- Windows 7 及之后版本（x86 / amd64 ）
- macos（amd64）

## 下载
预编译的二进制文件压缩包可在 [Github Release](https://github.com/libsgh/PanIndex/releases "release")下载，解压后方可使用。

## 安装

### 一键脚本
未完成
### 直接运行
启动参数<br>
-host=0.0.0.0 #绑定host，默认0.0.0.0<br>
-port=5238 #绑定端口号，默认5238<br>
-debug=false #调试模式，默认false<br>
-data_path=/path/to/data #数据目录（配置、目录信息、临时文件目录）<br>
-cert_file=/path/to/fullchain.pem # 开启ssl，证书文件<br>
-key_file=/path/to/privkey.pem # 开启ssl，证书文件密钥<br>
```bash
$ tar -xvf PanIndex-v1.0.0-linux-amd64.tar.gz
#nohup ./PanIndex -host=0.0.0.0 -port=5238 -debug=false > PanIndex.log &
$ nohup ./PanIndex > PanIndex.log &
```

### heroku部署

- 注册登录账号并绑卡，因为herokuku免费版有使用小时数限制，绑定信用卡可以使应用一直在线，**不扣费**
- 点击↓按钮跳转到heroku部署页面，修改 **CONFIG**

[![Deploy](https://www.herokucdn.com/deploy/button.png)](https://heroku.com/deploy?template=https://github.com/libsgh/PanIndex)

![](_images/1-2.png)
- 点击 **Deploy app** 完成部署

### docker部署
参考下面命令，映射`/app/data`目录到宿主机避免重启docker数据丢失！
```bash
docker pull iicm/pan-index:latest
docker stop PanIndex
docker rm PanIndex
docker run -itd \
 --name PanIndex \
 -d -p 5238:5238 \
 -v /home/single/data/docker/data/PanIndex/data:/app/data \
 -e PORT="5238" \
 iicm/pan-index:latest
```
### 从源码运行
- 安装git和golang
- 设置go环境变量`go env -w GO111MODULE=on`
- 如果是国内服务器，设置下代理`go env -w GOPROXY=https://goproxy.cn,direct`
```
$ git clone https://github.com/libsgh/PanIndex.git
$ cd PanIndex
$ nohup go run main.go > PanIndex.log &
```
也可以下载源码后自行编译成二进制程序再执行
以linux,amd64为例
```
$ CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o PanIndex
$ nohup ./PanIndex &
```
更多平台编译参考：[PanIndex-release-action](https://github.com/libsgh/PanIndex-release-action)
### 宝塔配置
### 视频教程