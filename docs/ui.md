> v1.0.4以上版本支持该功能， **由于版本更新经常涉及到页面调整，为方便升级，请尽量使用后台配置JS+CSS进行修改。**
- 在[release](https://github.com/libsgh/PanIndex/releases "release")处下载ui.zip包
- 解压到与PanIndex同级目录（或通过配置执行UI目录），按需修改templates及static目录中的内容
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
├── ├── ├── admin //系统配置
├── ├── ├── ├── index.html
├── ├── ├── ├── login.html
├── ├── ├── mdui
├── ├── ├── ├── index.html
├── ├── ├── bootstrap
├── ├── ├── ├── index.html
├── ├── ├── classic
├── ├── ├── ├── index.html
├── ├── ├── materialdesign
├── ├── ├── ├── index.html
```