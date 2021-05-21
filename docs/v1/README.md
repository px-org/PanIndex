# 快速开始
## 平台支持
由于PanIndex交叉编译需要cgo（sqlite），目前很多平台还不能很好的支持，如果你有特殊的编译需求，请告知我，我会尽量添加
- Linux （x86 / amd64 / arm / arm64 ）
- Windows 7 及之后版本（x86 / amd64 ）
- ~~macos（amd64）~~

## 下载PanIndex
预编译的二进制文件压缩包可在 [Github Release](https://github.com/libsgh/PanIndex/releases "release")下载，解压后方可使用。


## 直接运行
```bash
$ tar -xvf PanIndex-v1.0.0-linux-amd64.tar.gz
$ nohup ./PanIndex -config config.json > PanIndex.log &
```

## heroku安装

- 注册登录账号并绑卡，因为herokuku免费版有使用小时数限制，绑定信用卡可以使应用一直在线，**不扣费**
- 点击↓按钮跳转到heroku部署页面，修改 **CONFIG**

[![Deploy](https://www.herokucdn.com/deploy/button.png)](https://heroku.com/deploy?template=https://github.com/libsgh/PanIndex)

![](_images/1-2.png)
- 点击 **Deploy app** 完成部署

## docker安装
参考下面命令，需要注意配置依靠[环境变量](/#环境变量)，与配置文件的格式稍有不同，加密文件夹格式为`id:1234;id2:2345`。环境变量配置优先级高于`config.json`的配置。也可以挂载容器内的`/app`目录，在宿主机自定义`config.json`,`data`目录中是数据库文件
```bash
docker pull iicm/pan-index:latest
docker stop PanIndex
docker rm PanIndex
docker run -itd \
 --name PanIndex \
 -d -p 8080:8080 \
 -v /home/single/data/docker/data/PanIndex/data:/app/data \
 -e HOST="0.0.0.0" \
 ...
 iicm/pan-index:latest
```
## 从源码运行
安装git和golang
设置go环境变量go env -w GO111MODULE=on
如果是国内服务器，设置下代理go env -w GOPROXY=https://goproxy.cn,direct
从源码直接运行
```
$ git clone https://github.com/libsgh/PanIndex.git
$ cd PanIndex
$ go run main.go -config config/config.json
```
也可以下载源码后自行编译成二进制程序再执行
以linux,amd64为例
```
$ CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o PanIndex
$ nohup ./PanIndex -config config/config.json &
```
更多平台编译参考：[PanIndex-release-action](https://github.com/libsgh/PanIndex-release-action)

## 定时任务
- cookie刷新 `0 0 8 1/1 * ?`
- 目录缓存刷新 `0 0 0/1 * * ?`
- heroku防止休眠 `0 0/5 * * * ?`

## 接口
- 手动刷新登录cookie：`GET /api/refreshCookie?token=<ApiToken>`
- 手动刷新目录缓存：`GET /api/updateFolderCache?token=<ApiToken>`
- 分享链接直链解析：`GET /api/shareToDown?url=<分享链接>&fileId=<文件父级ID>&passCode=<访问密码>&subFileId=<文件ID，用于文件夹内的文件获取下载地址>`

> 如果分享的是文件，直接返回下载链接
> 如果分享链接是目录，返回文件目录信息（json格式），由于目录可能存在嵌套，fileId就是获取子目录的
> 下载分享目录内的文件，fileId传上级目录的文件ID，subFileId传该文件ID,文件ID可以从目录json中获取
> 访问密码有就传，没有就不传

# 配置项
## 配置文件

|  字段名         | 层级  | 必填  | 描述                                                    | 示例                           |
| :-------------: | :----:| :---: | :-----------------------------------------------------: | :----------------------------: |
|  host           | 0     | 否    | 服务监听地址                                            | 0.0.0.0                        |
|  port           | 0     | 是    | 端口                                                    | 8080                           |
|  mode           | 0     | 是    | 网盘模式：本地，天翼云，阿里teambition                                | native,cloud189(默认),teambition(项目或个人)    |
|  user           | 0     | 是    | 账号，一般是手机号或邮箱                                | 183xxxx7765，本地模式可不填                   |
|  password       | 0     | 是    | 密码                                          | 1234，本地模式可不填                           |
|  root_id        | 0     | 是    | 网盘根目录ID                                            | 天翼云：-11<br /> teambition-项目：项目id（正确的项目ID将自动开启）<br />teambition-个人：目录id<br />本地模式：绝对路径 |
|  pwd_dir_id     | 0     | 否    | 加密文件目录id和密码                                    | 数组                           |
|  id             | 1     | 否    | 加密目录id                                              | 5149xxx1353335，本地模式为绝对路径                 |
|  pwd            | 1     | 否    | 加密目录访问密码                                        | 1234                           |
|  hide_file_id   | 0     | 否    | 隐藏目录id ，多个文件`,`分隔                            | 213123,23445 【本地模式为绝对路径 】                  |
|  heroku_app_url | 0     | 否    | 部署后的herokuapp网盘地址，heroku部署必须               | https://app-name.herokuapp.com |
|  api_token      | 0     | 否    | 调用私有api的秘钥                                       | 1234                           |
|  theme          | 0     | 是    | 使用的主题，目前支持 classic, bootstrap, materialdesign, mdui | 默认为classic                      |
|  damagou        | 1     | 否    | 打码狗平台的用户名和密码，用于识别验证码                | username,password              |
|  only_Referer        | 0     | 否    | 简单防盗链，允许的 Referrer，留空为全部允许                | baidu.com              |
|  cron_exps      | 1     | 否    | 计划任务                                            | refresh_cookie,update_folder_cache,heroku_keep_alive              |
|  refresh_cookie      | 0     | 否    | 计划任务-刷新登录cookie                                            |        默认：`0 0 8 1/1 * ?` ，[cron在线生成](https://cron.qqe2.com/)      |
|  update_folder_cache      | 0     | 否    | 计划任务-刷新目录缓存                                            |      默认：`0 0 0/1 * * ?`，[cron在线生成](https://cron.qqe2.com/)       |
|  heroku_keep_alive      | 0     | 否    | 计划任务-heroku保持在线                                            |      默认： `0 0/5 * * * ?`，[cron在线生成](https://cron.qqe2.com/)       |

## config.json
```json
{
    "host": "0.0.0.0",
    "port": 8080,
    "mode": "cloud189",
    "user": "183xxxx7765",
    "password": "xxxx",
    "root_id": "-11",
    "pwd_dir_id": [
        {
            "id": "51496311321353335",
            "pwd": "1234"
        }
    ],
    "hide_file_id": "",
    "heroku_app_url":"https://pan-index.herokuapp.com",
    "api_token": "1234",
    "theme": "boot",
    "damagou": {
        "username":"",
        "password":""
    },
    "only_Referer": [],
    "cron_exps": {
        "refresh_cookie": "0 0 8 1/1 * ?",
        "update_folder_cache": "0 0 0/1 * * ?",
        "heroku_keep_alive": "0 0/5 * * * ?"
    }
}
```

## 环境变量
```bash
export HOST="0.0.0.0"
export MODE="cloud189"
export CLOUD_USER=""
export CLOUD_PASSWORD=""
export ROOT_ID="-11"
export PWD_DIR_ID="51496311321353335:1234"
export HIDE_FILE_ID=""
export API_TOKEN="1234"
export THEME="bootstrap"
export DMG_USER=""
export DMG_PASS=""
export ONLY_REFERxiaziaER="baidu.com,qq.com"
export CRON_REFRESH_COOKIE="0 0 8 1/1 * ?"
export CRON_UPDATE_FOLDER_CACHE="0 0 0/1 * * ?"
export CRON_HEROKU_KEEP_ALIVE="0 0/5 * * * ?"
```



# 文件预览

文件预览支持的格式：图片、音频、视频、office文档，其他不支持预览的文件，点击文件名将会进行下载

- 视频：h5播放兼容性最好的是mp4，其他的如mkv，可能播放会有问题

- 音频：支持封面和歌词，歌词必须是lrc格式，封面、歌词、音频必须同名（除后缀）

  ![](_images/audio.png)



# 自定义主题

> v1.0.4以上版本支持该功能
- 在[release](https://github.com/libsgh/PanIndex/releases "release")处下载ui.zip包
- 解压到与PanIndex同级目录，按需修改templates及static目录中的内容
- 模板中`{{}}`都是与go渲染html相关的表达式，请勿修改这部分内容
- 目录
```
├── PanIndex
├── static
├── ├── img       // 图片
├── ├── ├── favicon-cloud189.ico
├── ├── ├── favicon-native.ico
├── ├── ├── favicon-teambition.ico
├── ├── js       // js
├── ├── ├── main.js
├── templates
├── ├── pan
├── ├── ├── bootstrap
├── ├── ├── ├── index.html
├── ├── ├── classic
├── ├── ├── ├── index.html
├── ├── ├── materialdesign
├── ├── ├── ├── index.html
```

# 更多

## 常见问题

### 如何获取目录ID？
正常访问官方网盘页面，进入到你想分享目录的页面，浏览器里地址栏最后面的就是目录ID

若要使用teambition项目版：rootId，请使用**项目id**

![image-20210312180742254](_images/teambition-project.png)

**注**：当网盘启用本地模式，目录ID为分享目录的绝对路径，比如我想分享本地的`/opt`目录，`root_id`为`/opt`，密码目录同理

### PanIndex为何不使用前后端分离？
PanIndex的设计初衷是简单高效，前后分离会增加部署的复杂度，且会增加页面响应时间，PanIndex为单页面应用，目前页面的管理也相对容易，也为了更方便的适配heroku，目前将html，js，css，image都打包到二进制文件中，所以通常你只需要一个二进制文件加一个配置文件即可。

### 如何自定义页面？
参考[自定义主题](/#自定义主题)

### 登录失效、目录不更新该如何解决？
程序启动时会初始化登录状态，生成目录缓存数据。按照配置执行定时任务，默认登录cookie刷新天执行一次，目录缓存每小时更新一次，如果网盘文件不经常更新建议按天执行，经测试天翼云，teambition的cookie都比较稳定，所以cookie刷新也不建议频繁执行，频繁执行会导致验证登录。

计划任务修改请在配置中cron表达式，[cron在线生成](https://cron.qqe2.com/)
```
"refresh_cookie": "0 0 8 1/1 * ?",
"update_folder_cache": "0 0 0/1 * * ?",
"heroku_keep_alive": "0 0/5 * * * ?"
```
如果你更新了网盘文件想要立刻生效，可以手动调用[接口](#接口)接口进行刷新

### 天翼云出现验证码登录，如何解决？
由于天翼云盘登录会有一定几率触发验证码（1分钟内频繁登录），因此可以利用第三方验证码识别平台进行识别。

目前暂时使用的平台为打码狗 [damagou.top](http://www.damagou.top)，该平台需付费使用。

本功能默认关闭，如需启用只需在配置文件中填写平台用户名密码即可。

### 网盘使用经验

1. 天翼云网盘：普通版容量较小，15G空间，如果文件被多个不同ip访问下载，有一定几率触发限速，会员也无法幸免。
2. teambition-项目文件：个人测试速度限制在500K左右，不过容量暂时没有限制，也不需要内测资格即可使用。
3. teambition-个人文件：无限速，有使用空间限制，目前有2T空间，后期可能会合并到阿里云盘，不清楚会不会一直存在。
4. 本地目录：下载会占用服务器的带宽，而且服务器到期可能要面临文件的转移。
5. 阿里云盘：内测+teambition合并有3T，公测使用各种福利码可以达到8T，不过空间可能会回收，不确定是不是永久，阿里云一直都很套路，你懂得，据网友反馈，使用目录程序分享账号可能会ban，管控上比teambition更严格。

## 意见反馈

### Github Issue
- [PanIndex](https://github.com/libsgh/PanIndex/issues)

### Telegram 讨论组
- https://t.me/PanIndex

### QQ
- 359916450

## 捐助



<img src="_images/wechat.png" alt="wechat" style="zoom: 50%;" /><img src="_images/alipay.png" alt="wechat" style="zoom: 50%;" />