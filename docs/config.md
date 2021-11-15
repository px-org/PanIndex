### 注意事项
- 后台配置地址：`http://ip:port/?admin`
- 默认密码：`PanIndex`
- 第一次安装后需要进行配置， 请务必修改默认密码
- 部分配置需要重启生效
- 环境变量的优先级最高！挂载网盘后会自动刷新COOKIE和目录缓存，速度快慢取决于你服务器的网络以及你的文件数量
- 如果网盘中的文件没有更新，PanIndex也没必要更新缓存，不会影响目录访问及直链瞎子啊
- Heroku：后台配置好后，获取配置json，并复制到heroku新的环境变量`PAN_INDEX_CONFIG`中，端口号会根据环境变量`PORT`覆盖，无需关注

### 基础配置
> 右上角中间的按钮可以将完整配置生成JSON格式，方便环境变量导入。
* 绑定Host：默认`0.0.0.0`，修改后重启生效
* 绑定端口：默认`5238`，修改后重启生效
* 网站标题：默认为空，设置后将优先于网盘名称展示
* 首页账号切换
    * 默认账号：首页将显示默认账号，或顺序第一位的账号，`home`按钮切换
    * 全部账号：首页将以文件夹形式列出所有账号，`home`按钮依然可以切换
* 后台登录密码：默认`PanIndex`，注意保护隐私
* 接口 token：第一次安装时系统随机生成，注意保护隐私
* 排序：指定网盘目录列表的文件（夹）默认排序
### 外观
> 由于直接修改ui不方便后续升级，所以请尽量使用这里的配置修改主题、外观。
* 主题： mdui主题功能最全，也会长期更新，并且移动端友好
    * mdui（源自[JustList](https://github.com/txperl/JustList)）（跟随系统切换暗黑、明亮）
    * mdui-light（明亮模式）
    * mdui-dark（暗黑模式）
    * classic（经典主题，请勿使用，等待更新）
    * bootstrap（请勿使用，等待更新）
    * materialdesign（请勿使用，等待更新）
* 网站图标(Favicon)：网站图标Url，为空将使用系统默认图标。根据网盘不同默认图标也不同。
* 自定义底部信息(Footer)：可以在此处修改备案信息，及站长相关链接，支持html代码片段。
  ```html
   ©2021 <a href="https://github.com/libsgh" target="_blank">libsgh</a>. All rights reserved.
  ```
* 自定义CSS：定义在页面head部分的html片段，可以是style，也可以指定外部css样式表。
    ```html
    <style>
    ...
    </style>
    ```
* 自定义Javascript：定义在页面底部的html代码片段，可以是script，也可以引用外部Javascript，支持绝大部分jquery语法。
    ```html
    <script>
    alert("测试");
    </script>
    ```
### 文件（夹）加密
* 加密的文件（夹）将无法被搜索到，文件ID，为文件在网盘中的ID，参考[文件ID获取方法](https://libsgh.github.io/PanIndex/#/question?id=%e5%a6%82%e4%bd%95%e8%8e%b7%e5%8f%96%e7%9b%ae%e5%bd%95id%ef%bc%9f)
### 隐藏文件（夹）
* 同上
### 防盗链
* 许可域名：允许的 Referrer，多个逗号分隔，例：`baidu.com,google.com`，不包含http请求头。
* 启用防盗链
* 允许空Referrer请求：当开启防盗链时可能会影响下载，注意勾选此选项。

### 网盘挂载
- 显示名称：会修改网页标题，每个账号可不一致
- 网盘模式
    - Native： 本地文件系统，服务器某一目录的文件列表，因为实时获取所以无需更新cookie和目录缓存
    - FTP： 服务器地址格式为`ip:port`，例如`192.169.1.1:21`
    - WebDav：服务器地址格式为http请求地址，例如`https://webdav.mydomain.me`
    - Cloud189：天翼云网盘，基于用户名（手机号）和密码，程序自动检测cookie是否有效并自动刷新。
    - Teambition：阿里teambition项目盘，根目录ID需要填写项目ID。
    - Teambition国际版：阿里teambition国际盘，目前只有项目文件，目录ID为项目ID
    - 和彩云：由于登录会有验证码，所以不会采用自动登录的方式，请登录网页版复制完整COOKIE并输入手机号。
    - Aliyundrive：阿里云盘，需要填入**手机端的**的`refresh_token`，推荐使用 [PanIndex Tool](https://mgaa.noki.workers.dev/) 扫码获取。
    - OneDrive、世纪互联：按照[PanIndex Tool](https://mgaa.noki.workers.dev/) 教程指引获取授权信息。
    - GoogleDrive：按照[PanIndex Tool](https://mgaa.noki.workers.dev/) 教程指引获取授权信息，请注意此模式下有一些限制：下载文件需要登录google，无法预览视频（跨域），服务器需要特殊网络环境等。
    > 由于阿里云的`refresh_token`和`access_token`有效期为2小时，第一次填入后，系统会自动刷新，所以`refresh_token`值会变，但是可以保持有效。
- 根目录ID(路径)：native、webdav、ftp的ID格式为对应路径，teambition为项目ID，其他为目录ID，[如何获取？](https://libsgh.github.io/PanIndex/#/question?id=%e5%a6%82%e4%bd%95%e8%8e%b7%e5%8f%96%e7%9b%ae%e5%bd%95id%ef%bc%9f)
  
  > 这里填写你要分享的目录ID，如果你想分享网盘的根目录，天翼云为`-11`，阿里云盘为`root`，合彩云盘为`00019700101000000001`
- 定时刷新缓存：为空将关闭缓存，请不要过于频繁的定时刷新，建议间隔频率1小时以上。
- 定时缓存目录：指定定时刷新PanIndex的虚拟目录，设置经常更新的目录，可以提高缓存效率。
- 缓存是否包含子目录：取消勾选后将不会递归更新子目录。
- 文件上传：选中需要上传的文件并指定目录，主要是用于几十M以内的小文件的上传。高级文件管理操作推荐使用官方客户端。
- 刷新令牌：当登录状态不正常时，可以手动进行刷新。
- 刷新缓存：当网盘文件有更新时可以手动刷新目录缓存。

**当文件有重复时，可以修改保存网盘，该操作会刷新登录状态并重新缓存**

### 环境变量

环境变量主要用于docker（docker）部署场景，vps下无需关注。另外，环境变量优先级最高。

| 变量名称            | 变量值     | 描述                                                     |
| ------------------- | ---------- | -------------------------------------------------------- |
| PAN_INDEX_CONFIG    | -          | 完整配置文件，可从后台获取                               |
| PAN_INDEX_DEBUG     | true/false | 是否开启调试模式，debug模式将输出更多日志，方便问题追踪  |
| PAN_INDEX_DATA_PATH | /opt/data  | 数据目录，默认与程序同级`data`目录下                     |
| PORT                | -          | 启动端口号，由于Heroku端口号随机，并需要从此环境变量获取 |

