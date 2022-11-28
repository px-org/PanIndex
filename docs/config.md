### 注意事项
- 后台地址（默认）：`http://ip:port/admin`
- 默认账号：`admin`
- 默认密码：`PanIndex`
- 第一次安装后需要进行配置， 请务必修改默认账号、密码
- 部分配置需要重启生效
## 通用
### 基础配置
* 网站标题：默认为空，设置后将优先于网盘名称展示
* 网站路径前缀：默认为空，如需域名多层级目录反代，可设置此项，设置后同步修改nginx，格式：`/file`
```
location /file/ {
   proxy_set_header Host $host;
   proxy_set_header X-Real-IP $remote_addr;
   proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
   proxy_pass http://127.0.0.1:5238/;
   client_max_body_size    1000m;
}
```
* 后台管理地址：默认为`/admin`
* 首页账号切换
    * 默认账号：首页将显示默认账号，或顺序第一位的账号，`home`按钮切换
    * 全部账号：首页将以文件夹形式列出所有账号，`home`按钮依然可以切换
* 静态资源cdn
* 登录账号：默认`admin`，请及时修改默认账号
* 登录密码：默认`PanIndex`，请及时修改默认密码
* 排序：指定网盘目录列表的文件（夹）默认排序

### 外观
> 由于直接修改ui不方便后续升级，所以请尽量使用这里的配置修改主题、外观。
* 主题： mdui主题功能最全，也会长期更新，并且移动端友好
    * mdui（源自[JustList](https://github.com/txperl/JustList)）（跟随系统切换暗黑、明亮）
    * mdui-light（明亮模式）
    * mdui-dark（暗黑模式）
    * classic（经典主题）
    * bootstrap
* 网站图标(Favicon)：网站图标Url，为空将使用系统默认图标。根据网盘不同默认图标也不同。
* 自定义底部信息(Footer)：可以在此处修改备案信息，及站长相关链接，支持html代码片段。
  ```html
   ©2022 <a href="https://github.com/libsgh" target="_blank">libsgh</a>. All rights reserved.
  ```
  * 借助于Footer添加评论系统
    * [valine](https://valine.js.org/)
    ```html
     <script src='//unpkg.com/valine/dist/Valine.min.js'></script>
     <div id="vcomments" class="mdui-p-a-2"></div>
     <script>
        new Valine({
          el: '#vcomments',
          appId: '',
          appKey: ''
        })
     </script>
     ©2021 <a href="https://github.com/libsgh/PanIndex" target="_blank">PanIndex</a>. All rights reserved.
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
### 文件预览  
* 是否开启预览：当关闭时访问文件链接将进行下载而不是跳转预览页面。
* 配置预览文件的后缀名，推荐使用默认选项。

## 安全
### 访问控制
* 短链行为：访问文件短链接是下载还是跳转文件预览页
* 访问控制模式
  - 公开：自由访问任何未加密、未隐藏的目录、文件
  - 仅直链：仅可访问文件下载直链，访问目录会返回404，登录后可显示文件列表
  - 直链+预览：可访问文件下载直链、及文件预览页面，访问目录也会返回404，登录后可显示文件列表
  - 登录：访问任意文件、目录均需要登录，登录后可以正常访问目录、文件

### 文件（夹）加密
* 文件路径为PanIndex的虚拟路径，即请求文件的url路径，例如：`https://localhost:5238/ali/a.txt`，加密路径为`/ali/a.txt`。
* 注意：网盘名称修改可能导致加密失效。
* 同一路径可以设置多个访问密码，添加时密码留空将随机生成。
* 密码文件格式：每行一个，格式`路径  密码  有效期  备注`，制表符（\t）分隔。
* 分流名称同理。

### 隐藏文件（夹）
* 同文件加密。

### 防盗链
* 许可域名：允许的 Referrer，多个逗号分隔，例：`baidu.com,google.com`，不包含http请求头，3.x版本开始支持正则。
* 启用防盗链
* 允许空Referrer请求：当开启防盗链时可能会影响下载，注意勾选此选项。

## 网盘挂载
* 添加、修改网盘
  - 网盘名称：将用于区分网盘文件的首位路径
  - 网盘模式
    - Native： 本地文件系统，服务器某一目录的文件列表，因为实时获取所以无需更新cookie和目录缓存
    - FTP： 服务器地址格式为`ip:port`，例如`192.169.1.1:21`
    - WebDav：服务器地址格式为http请求地址，例如`https://webdav.mydomain.me`
    - Cloud189：天翼云网盘，基于用户名（手机号）和密码，程序自动检测cookie是否有效并自动刷新。
    - Teambition：阿里teambition项目盘，根目录ID需要填写项目ID。
    - Teambition国际服：阿里teambition国际盘，目前只有项目文件，目录ID为项目ID
    - 和彩云：由于登录会有验证码，所以不会采用自动登录的方式，请登录网页版复制完整COOKIE并输入手机号，COOKIE有效期为一个月。
    - Aliyundrive：阿里云盘，需要填入**手机端的**的`refresh_token`，点击文本框下方的扫码获取链接。
    - OneDrive、世纪互联：按照[PanIndex Tool](https://pt.noki.icu/) 教程指引获取授权信息，其中设置网站ID可挂载SharePoint挂载。
    - GoogleDrive：按照[PanIndex Tool](https://pt.noki.icu/) 教程指引获取授权信息，请务必勾选流量中转，服务器需要特殊网络环境。
    - S3：基于AWS S3的SDK实现，已通过测试的存储：aws S3、阿里云OSS（virtual hosted）、腾讯云COS、Oracle object-storage，注意配置公开访问权限及跨域设置，根目录ID设置：存储桶根目录留空，子目录格式：test/abc/。
    - PakPik：配置登录邮箱及密码，挂载全局根目录需设置目录ID为空
    > 由于阿里云的`refresh_token`和`access_token`有效期为2小时，第一次填入后，系统会自动刷新，所以`refresh_token`值会变，但是可以保持有效。
    - 根目录ID(路径)：native、webdav、ftp的ID格式为目录的绝对路径，teambition请分别输入项目ID和目录ID，[如何获取？](https://libsgh.github.io/PanIndex/#/question?id=%e5%a6%82%e4%bd%95%e8%8e%b7%e5%8f%96%e7%9b%ae%e5%bd%95id%ef%bc%9f)
  
    > 这里填写你要分享的目录ID，如果你想分享网盘的根目录，天翼云为`-11`，阿里云盘为`root`，和彩云盘为`00019700101000000001`，S3留空
    - 流量中转：中转地址为空将使用本机中转，设置的CDN域名会替换下载地址中的域名。
  - 缓存设置
    - 缓存策略：API直连（v3），除几种本地模式外，其他盘仅用于测试，命中缓存（v3）：当访问目录时进行内存级缓存，超过有效期将自动失效，完全缓存：缓存所有文件到数据库，适用于文件不多，修改不频繁的场景。
    - 定时刷新缓存（DB Cache）：为空将关闭缓存，请不要过于频繁的定时刷新，建议间隔频率1小时以上。
    - 定时缓存目录（DB Cache）：指定定时刷新PanIndex的虚拟目录，设置经常更新的目录，可以提高缓存效率，当网盘开启分流时，这里要填写分流路径。
    - 缓存是否包含子目录（DB Cache）：取消勾选后将不会递归更新子目录。
  - 文件上传：选中需要上传的文件并指定目录，主要是用于几十M以内的小文件的上传，上传会消耗服务器流量。高级文件管理操作推荐使用官方客户端。
  - 刷新令牌：当登录状态不正常时，可以手动进行刷新。
  - 刷新缓存：当网盘文件有更新时可以手动刷新目录缓存。
  - 删除：删除不再使用的网盘

**当文件有重复时，可以修改缓存策略，该操作会重置缓存，仅DB缓存的文件可以被搜索到**

## 分流下载
创建分流下载，选中的网盘将以轮询的方式访问目录、下载，同时账号列表也将合并成一个。
## 缓存管理
无论是内存缓存还是DB缓存都可以在这里搜索到，并查看删除
## WebDav
- 是否开启
- WebDav的请求路径
- 访问权限：只读，读写
- 下载行为：由于WebDav对于302跳转直链下载并不能很好的支持，可以选择中转方式，建议在个人电脑进行使用。
- 用户名
- 密码

