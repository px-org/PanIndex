## 配置说明
### 注意事项
- 后台配置地址：`http://ip:port/?admin`
- 默认密码：`PanIndex`
- 第一次安装后需要进行配置， 请务必修改默认密码
- 部分配置需要重启生效
- 绑定账号后会自动刷新COOKIE和目录缓存，速度快慢取决于你服务器的网络以及你的文件数量
- 可以通过缓存记录查看缓存结果，如果失败，也可以手动同步
> 请不要频繁刷新以免出现验证登录
    
![](_images/cache-record.png)
- 默认关闭目录缓存定时任务，如有需要请自行设置，heroku每天至少执行一次任务，建议`corn：0 0 4 1/1 * ?`

![](_images/cron.png)

- Heroku：后台配置好后，获取配置json，并复制到heroku新的环境变量`PAN_INDEX_CONFIG`中，端口号会根据环境变量`PORT`覆盖，无需关注

### 基础配置
* 绑定Host：默认`0.0.0.0`，修改后重启生效
* 绑定端口：默认`5238`，修改后重启生效
* 主题： mdui主题功能最全，也会长期更新，并且移动端友好
  * mdui（跟随系统切换暗黑、明亮）
  * mdui-light（明亮模式）
  * mdui-dark（暗黑模式）
  * classic（经典主题，不支持账号前端切换及搜索，适用于单账号）
  * bootstrap
  * materialdesign
* 后台登录密码：默认`PanIndex`，注意保护隐私
* 接口 token：第一次安装时系统随机生成，注意保护隐私
* 密码文件（夹）：格式`id1:pwd1,path1:pwd2`
* 隐藏文件ID（路径）:id1,path1
* 防盗链：允许的 Referrer，多个逗号分隔，例：`baidu.com,google.com`
* 自定义网站图标链接
 * 可以将自定义图标`favicon.ico`上传至网盘，填入图片直链
* 自定义底部信息

```html
©2021 <a href="https://github.com/libsgh" target="_blank">libsgh</a>. All rights reserved.
```
* 底部查看完整配置，用于那些沙盒容器平台配置环境变量

### 账号绑定
* 显示名称：会修改网页标题，每个账号可不一致
* 网盘模式
 * native： 本地模式，服务器某一目录的文件列表，因为实时获取所以无需更新cookie和目录缓存
 * cloud189：天翼云网盘
 * teambition：阿里teambition盘，包括个人网盘和项目文件，依据根目录ID设定自动判断
 * teambitionguo国际版：阿里teambition国际盘，目前只有项目文件，目录ID为项目ID
* 用户名：部分模式必需，一般是手机号或邮箱
* 密码
* 根目录ID(路径)：native为绝对路径，teambition为项目ID，其他为目录ID

### 文件上传
* 手动上传
 * 支持多文件上传，远程目录请填写网盘的相对路径，例如：
 我想上传1这个目录，就填写/1
 ![](_images/upload-remote-dir.jpg)
* 自动同步（待实现）
