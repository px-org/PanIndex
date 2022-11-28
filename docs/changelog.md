### [v3.1.1](https://github.com/libsgh/PanIndex/compare/v3.1.0...v3.1.1) (2022-11-28)
> 温馨提示：该版本涉及页面改动，升级时需要使用新版本ui，对于页面的个性化调整，尽量使用js或css，而不是直接修改页面，这样方便后期升级！

##### Bug Fixes

- 修复classic、bootstrap主题的一些错误
- 完全缓存模式下阿里云盘API访问频繁导致的目录信息丢失
- 修复某些情况下UI目录找不到的问题

##### Improve
- 优化cdn资源访问
- Mdui暗黑主题显示优化（去除中间黑框）
- 字体图标改为本地访问

##### Features

- [配置管理]增加配置-网站路径前缀，用于nginx反代非根目录跳转，配置`/file` `http://noki.icu/file/`，参考[nginx配置](/config.md?id=基础配置)
- 后台账号、密码重置，命令更新`-reset_password=1234`、`-reset_user=admin`，至此有[三种方法](/question.md?id=忘记后台登录账号、密码如何解决？)解决账号密码忘记问题
- 启动配置（config.json、Env、Program Args）增加参数`ui`，用于自定义UI目录，默认将从执行程序所在目录查找
- 更新Go构建版本1.19

### [v3.1.0](https://github.com/libsgh/PanIndex/compare/v3.0.9...v3.1.0) (2022-08-24)

##### Bug Fixes

- 修复由跨域设置导致的部分WebDav连接失败
- [PikPak]定时刷新过期TOKEN
- 增加WebDav文档
- 分流添加页面高度调整


### [v3.0.9](https://github.com/libsgh/PanIndex/compare/v3.0.8...v3.0.9) (2022-08-23)

##### Bug Fixes

- 修复文件扩展名为大写时文件类型识别错误 by [wshon](https://github.com/wshon)
- 修复跨域访问错误
- 新增网盘时排序错误
- 批量缓存无法并发执行

##### Improve
- 提高WebDav访问兼容性
- 优化暗黑主题显示
- 更新播放器artplayer

##### Features

- [配置管理]增加配置-访问控制
  - 公开：自由访问任何未加密、未隐藏的目录、文件
  - 仅直链：仅可访问文件下载直链，访问目录会返回404，登录后可显示文件列表
  - 直链+预览：可访问文件下载直链、及文件预览页面，访问目录也会返回404，登录后可显示文件列表
  - 登录：访问任意文件、目录均需要登录，登录后可以正常访问目录、文件

- [网盘新增]S3、PikPak

- [配置管理]新增登录账号：默认：admin

- [配置管理]管理接口刷新token

- [配置管理]优化密码配置，有效期、备注、随机密码、导入、生成密码文件短链

- [配置管理]网盘挂载新增备注字段

- 下载直链支持密码参数，可以直接下载密码文件，http://localhost:5238/a/b.txt?pwd=1234

- 支持密码重置，启动时指定参数`-reset_password=1234`

- 增加404页面

- 基于dev分支的自动构建版本，[nightly](https://nightly.link/libsgh/PanIndex/workflows/nightly-build/dev)

- 基于dev分支的docker镜像：iicm/pan-index:dev


### [v3.0.8](https://github.com/libsgh/PanIndex/compare/v3.0.7...v3.0.8) (2022-05-08)

##### Bug Fixes

- 天翼云中转无法下载
- 配置导入账号部分配置丢失

##### Features

- 支持Sharepoint挂载
- 更新js依赖及部分npm cdn来源

#### Dev Plan

- 纯直链系统（屏蔽、目录访问、文件预览）
- 新增网盘支持
- 分页、搜索优化
- 播放增强

### [v3.0.7](https://github.com/libsgh/PanIndex/compare/v3.0.6...v3.0.7) (2022-04-19)

##### Bug Fixes

- 刷新全部缓存丢失文件
- 页面工具栏菜单展示不全
- 【Teambition】项目盘项目ID无法保存
- 中转域名保修改无效

##### Features

- 配置支持导入导出
- HEAD.md、README.md渲染优化
- 预览页去除ID信息

### [v3.0.6](https://github.com/libsgh/PanIndex/compare/v3.0.0...v3.0.6) (2022-04-01)

##### Bug Fixes

- 手机端视频无法播放
- 音频文件无法播放
- 分流列表展示不全
- UI文件本地修改无效
- 密码文件下载直链访问控制

##### Features

- 一键缓存全部账号
- 网盘HOST绑定，文件列表过滤
- 回车输入提交访问密码
- 配置项：`README.md`、`HEADME.md`增加渲染开关
- 配置项：音频、视频字幕加载、弹幕库支持
- 兼容2.x文件直链访问，`/d_0/xx`

### [v3.0.0](https://github.com/libsgh/PanIndex/compare/v2.0.9...v3.0.0) (2022-03-04)

#### 注意：由于新版改动较大，建议重新安装，此为V3测试版本，如遇BUG请及时反馈

1. 增加缓存策略（No Cache、Memory、Database）
2. 完善各类网盘操作api（文件夹创建、复制、移动、删除等）
3. 优化目录缓存慢
4. 分流下载
5. 某些浏览器 阿里云盘无法下载的问题
6. 定时任务错乱
7. 本地模式支持分片下载
8. 提供WebDav服务
9. 略缩图布局切换
10. 自定义后台地址
11. windows运行隐藏窗口
12. mdui主题视频播放支持播放列表（自然排序）
13. 多数据源支持：mysql、sqlite、postgres
14. 静态资源CDN配置
15. classic、bootstrap主题修复
16. 密码目录访问优化
17. 流量中转

### [v2.0.9](https://github.com/libsgh/PanIndex/compare/v2.0.8...v2.0.9) (2021-12-20)

##### Bug Fixes

**此版本大概是2.x系列最后一个版本，主要是紧急应对jsdelivr在国内无法访问的问题**

- 修改mdui主题及后台引用的js、css从本地读取

### [v2.0.8](https://github.com/libsgh/PanIndex/compare/v2.0.7...v2.0.8) (2021-11-15)

##### Bug Fixes

- 修复文件（夹）加密无法访问
- 『阿里云盘』增加转码播放切换按钮（阿里云盘转码地址15分钟有效期，会导致播放中断，尽量用原地址播放）
- 『阿里云盘』延长下载地址有效期
- 修复一些后台显示错误
- 修复markdown文件渲染失败
- 修复Heroku构建失败

##### Features

- 新增『FTP』『WebDav』『OneDrive世纪互联』『和彩云』『谷歌云盘』五种网盘模式

### [v2.0.7](https://github.com/libsgh/PanIndex/compare/v2.0.6...v2.0.7) (2021-10-14)

**为方便以后配置扩展，此版本主要是对后台配置进行了重写，另外只有mdui主题做了功能增强，请尽量使用该主题**

##### Bug Fixes

- 后台配置优化：自定义账号、文件列表排序、防盗链、隐藏&密码文件
- 修复『天翼云盘』上传接口、取消文件夹下载
- 解决列表页图片预览加载优化 
- 优化密码输入、主题切换记住功能 
- 去除无用的定时任务
- 预览文件后缀可配置
- 『本地模式』打包下载临时文件清理

##### Features

- 添加分享短链&二维码
- 添加下载链接一键提取
- 『阿里云盘』视频支持转码播放
- pdf预览支持手机端
- 『阿里云盘』支持文件夹下载
- 支持epub格式的电子书预览
- 添加分享短链&二维码
- 添加下载链接一键提取
- 『阿里云盘』视频支持转码播放
- 列表页支持`HEAD.md`渲染

### [v2.0.6](https://github.com/libsgh/PanIndex/compare/v2.0.5...v2.0.6) (2021-09-27)

**此版本是天翼云紧急修复版本，原计划的新版后台配置做了优化增强，改动较大，目前还在开发中，`dev`分支处于不可部署状态，由于最近事情较多，进度有点慢，大家耐心等待。**

##### Bug Fixes

- 首页`README.md`增加缓存 
- 修复『天翼云盘』缓存接口失效 
- 处理GetFileData文件找不到异常 
- 解决关闭浏览器，个性化设置丢失问题

### [v2.0.5](https://github.com/libsgh/PanIndex/compare/v2.0.4...v2.0.5) (2021-08-26)

**若升级后密码文件夹无法访问，请删除cookie重试**

##### Bug Fixes

- 解决访问空目录无法打开的问题
- 「本地模式」windows下路径错误 
- 解决主题失效问题
- 「OneDrive」修复只能缓存获取200个文件的问题
- 搜索、排序、密码功能增强

##### Features

- 「文件预览」上一个/下一个过滤隐藏文件
- 添加「返回顶部」按钮
- 支持自定义后台页面
- 搜索改为全局搜索
- 详情页显示文件ID
- 「后台」增加配置项、首页账号切换、网站标题
- 预览优化：视频播放支持webvtt字幕、音频播放支持播放列表、pdf、markdown预览

### [v2.0.4](https://github.com/libsgh/PanIndex/compare/v2.0.3...v2.0.4) (2021-07-29)

##### Bug Fixes

- **修复天翼云登录失败** by [Akkariin Meiko](https://github.com/kasuganosoras)
- 解决多线程下载导致的内存溢出
- 解决OneDrive个人版无法同步子目录的问题
- 解决README获取下载地址异常
- OneDrive刷新缓存的过程中，`refresh_token`失效导致的文件丢失

##### Features

- mdui主题增加预览详情页
- 增加通过命令获取配置参数，例：`-cq=port`
- 定时缓存可以指定某一目录，提高缓存效率
- 优化账号绑定提示
- 本地模式文件类型优化
- 视频播放支持m3u8

### [v2.0.3](https://github.com/libsgh/PanIndex/compare/v2.0.2...v2.0.3) (2021-07-02)

##### Bug Fixes

* 解决环境变量配置DEBUG模式不打印SQL语句的问题 [8c8fdfb](https://github.com/libsgh/PanIndex/commit/8c8fdfb7e1de7142ae83679dad69fac93cd070e8)
* 优化检测逻辑，避免一直错误登录 [3c802b8](https://github.com/libsgh/PanIndex/commit/3c802b8a049729d44bfbb3c65ffadc4c3f2ba87c)

##### Features

* 添加对Onedrive的支持 ([#42](https://github.com/libsgh/PanIndex/issues/42)) ([9098c02](https://github.com/libsgh/PanIndex/commit/9098c0263d0ae541f9d3b86e9775e8216845d132))