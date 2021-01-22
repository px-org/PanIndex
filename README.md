# PanIndex
[![Deploy](https://www.herokucdn.com/deploy/button.png)](https://heroku.com/deploy?template=https://github.com/libsgh/PanIndex)

简易的天翼云网盘目录列表
- [x] cookie方式
- [x] 下载直链
- [x] 文件夹打包下载
- [x] 图片、视频、office文件在线预览
- [x] 定时缓存文件及目录信息（每小时同步一次）
- [x] 定时同步cookie（每天8点）
- [x] heroku防止休眠（每5分钟访问一次应用）

demo: [https://pan-index.herokuapp.com](https://pan-index.herokuapp.com "https://pan-index.herokuapp.com"),[https://pan.noki.top](https://pan.noki.top "https://pan.noki.top")
## 一键部署到heroku
1.  注册登录账号并绑卡，因为herokuku免费版有使用小时数限制，绑定信用卡可以使应用一直在线，**不扣费**
2.  点击上面↑的按钮，跳转到heroku部署页面，修改 **CONFIG**
![](https://raw.githubusercontent.com/libsgh/PanIndex/master/doc/1-2.png)
3. 点击 **Deploy app** 完成部署

## docker部署
参考下面命令，需要注意配置依靠环境变量，与配置文件的格式稍有不同，加密文件夹格式为`id:1234;id2:2345`。环境变量配置优先级高于`config.json`的配置。也可以挂载容器内的`/app`目录，在宿主机自定义`config.json`,`data`目录中是数据库文件
```
docker pull iicm/pan-index:latest
docker stop PanIndex
docker rm PanIndex
docker run -itd \
 --name PanIndex \
 -d -p 8080:8080 \
 -v /home/single/data/docker/data/PanIndex/data:/app/data \
 -e HOST="0.0.0.0" \
 -e MODE="cloud189" \
 -e CLOUD_USER="1860****837" \
 -e CLOUD_PASSWORD="1234" \
 -e ROOT_ID="-11" \
 -e PWD_DIR_ID="51496311321353335:1234" \
 -e HIDE_FILE_ID="" \
 -e HEROKU_APP_URL="" \
 -e API_TOKEN="" \
 -e THEME="bootstrap" \
 -e DMG_USER="" \
 -e DMG_PASS="" \
 -e CRON_REFRESH_COOKIE="0 0 8 1/1 * ?" \
 -e CRON_UPDATE_FOLDER_CACHE="0 0 0/1 * * ?" \
 -e CRON_HEROKU_KEEP_ALIVE="0 0/5 * * * ?" \
 iicm/pan-index
```
## 程序包独立部署（vps）
在[release](https://github.com/libsgh/PanIndex/releases "release")处下载PanIndex-xxx.tar.gz包
```bash
$ tar -xvf PanIndex-v1.0.0-linux-amd64.tar.gz
$ ./PanIndex -config.path=config.json
```

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
  

## 验证码识别
由于天翼云盘登录会有一定几率触发验证码，因此可以利用第三方验证码识别平台进行识别。

目前暂时使用的平台为打码狗 [damagou.top](http://www.damagou.top)，该平台需付费使用。

本功能默认关闭，如需启用只需在配置文件中填写平台用户名密码即可。

**配置文件说明:**

|  字段名         | 层级  | 必填  | 描述                                                    | 示例                           |
| :-------------: | :----:| :---: | :-----------------------------------------------------: | :----------------------------: |
|  host           | 0     | 否    | 服务监听地址                                            | 0.0.0.0                        |
|  port           | 0     | 是    | 端口                                                    | 8080                           |
|  mode           | 0     | 是    | 网盘模式：本地，天翼云，阿里teambition                                | native,cloud189(默认),teambition                    |
|  user           | 0     | 是    | 天翼云账号，一般是手机号                                | 183xxxx7765                    |
|  password       | 0     | 是    | 天翼云账号密码                                          | 1234                           |
|  root_id        | 0     | 是    | 网盘根目录ID                                            | -11，代表天翼云顶层目录        |
|  pwd_dir_id     | 0     | 否    | 加密文件目录id和密码                                    | 数组                           |
|  id             | 1     | 否    | 加密目录id                                              | 5149xxx1353335                 |
|  pwd            | 1     | 否    | 加密目录访问密码                                        | 1234                           |
|  hide_file_id   | 0     | 否    | 隐藏目录id ，多个文件`,`分隔                            | 213123,23445                   |
|  heroku_app_url | 0     | 否    | 部署后的herokuapp网盘地址，heroku部署必须               | https://app-name.herokuapp.com |
|  api_token      | 0     | 否    | 调用私有api的秘钥                                       | 1234                           |
|  theme          | 0     | 是    | 使用的主题，目前支持 classic, bootstrap, materialdesign | bootstrap                      |
|  damagou        | 1     | 否    | 打码狗平台的用户名和密码，用于识别验证码                | username,password              |
|  cron_exps      | 1     | 否    | 计划任务                                            | refresh_cookie,update_folder_cache,heroku_keep_alive              |
|  refresh_cookie      | 0     | 否    | 计划任务-刷新登录cookie                                            |        默认：`0 0 8 1/1 * ?` ，[cron在线生成](https://cron.qqe2.com/)      |
|  update_folder_cache      | 0     | 否    | 计划任务-刷新目录缓存                                            |      默认：`0 0 0/1 * * ?`，[cron在线生成](https://cron.qqe2.com/)       |
|  heroku_keep_alive      | 0     | 否    | 计划任务-heroku保持在线                                            |      默认： `0 0/5 * * * ?`，[cron在线生成](https://cron.qqe2.com/)       |

config.json
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
    "theme": "bootstrap",
    "damagou": {
        "username":"",
        "password":""
    },
    "cron_exps": {
        "refresh_cookie": "0 0 8 1/1 * ?",
        "update_folder_cache": "0 0 0/1 * * ?",
        "heroku_keep_alive": "0 0/5 * * * ?"
    }
}
```

