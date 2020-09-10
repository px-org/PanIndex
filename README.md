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

demo: https://pan-index.herokuapp.com
## 一键部署到heroku
1.  注册登录账号并绑卡，因为herokuku免费版有使用小时数限制，绑定信用卡可以使应用一直在线，**不扣费**
2.  点击上面↑的按钮，跳转到heroku部署页面，修改 **CONFIG**
![](https://raw.githubusercontent.com/libsgh/PanIndex/master/doc/1-2.png)
3. 点击 **Deploy app** 完成部署

## docker部署

## 程序包独立部署（windows or linux）

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
    "heroku_app_url":"https://pan-index.herokuapp.com"
}
```