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