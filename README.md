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
 -e CLOUD_USER="1860****837" \
 -e CLOUD_PASSWORD="1234" \
 -e ROOT_ID="-11" \
 -e PWD_DIR_ID="51496311321353335:1234" \
 -e HIDE_FILE_ID="" \
 -e HEROKU_APP_URL="" \
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
- 手动刷新目录缓存：`GET /api/updateFolderCache?token=<ApiToken>`


**配置文件说明:**

|  字段名 |  层级  | 必填  | 描述  | 示例  |
| :------------: | :------------:| :------------: | :------------: | :------------: |
|  port | 0  | 是 | 端口  | 8080  |
|  user | 0  | 是 | 天翼云账号，一般是手机号  | 183xxxx7765  |
|  password | 0  | 是 |  天翼云账号密码 |  1234 |
|  root_id |  0 | 是 | 网盘根目录ID  |  -11，代表天翼云顶层目录 |
| pwd_dir_id  | 0  | 否 | 加密文件目录id和密码  | 数组  |
| id  | 1  | 否 |  加密目录id |  5149xxx1353335 |
|  pwd |  1 | 否 |  加密目录访问密码 | 1234  |
|  hide_file_id |  0 | 否 |  隐藏目录id ，多个文件`,`分隔 | 213123,23445  |
|  heroku_app_url |  0 | 否 | 部署后的herokuapp网盘地址，heroku部署必须 | https://app-name.herokuapp.com  |
|  api_token |  0 | 否 | 调用私有api的秘钥 | 1234  |

config.json
```json
{
    "port": 8080,
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
    "api_token": "1234"
}
```