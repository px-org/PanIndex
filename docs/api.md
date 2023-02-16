# API

> Since v3.1.2

## 公共

### 1. 获取API程序信息

| URL                 | Request Method | Content-Type                      |
| ------------------- | -------------- | --------------------------------- |
| /api/v3/public/info | GET            | application/x-www-form-urlencoded |

#### 请求参数

无

#### 返回参数

| Name       | Type   | Description    |
| ---------- | ------ | -------------- |
| author     | string | 程序作者       |
| commit_sha | string | git commit sha |
| name       | string | 程序名称       |
| version    | string | 版本号         |

#### 返回示例

```json
{
    "data": {
        "author": "Libs",
        "commit_sha": "f0de26b2e8649a944398c9f504d24122ed71d5a1",
        "name": "PanIndex",
        "version": "v3.1.2"
    },
    "msg": "success",
    "status": 0
}
```



### 2. 获取配置信息
sho
| URL                        | Request Method | Content-Type                      |
| -------------------------- | -------------- | --------------------------------- |
| /api/v3/public/config.json | GET            | application/x-www-form-urlencoded |

#### 请求参数

无

#### 返回参数

| Name           | Type   | Description            |
| -------------- | ------ | ---------------------- |
| account_choose | string | 账号选择方式           |
| audio          | string | 音频文件预览后缀       |
| css            | string | 自定义css              |
| doc            | string | office文档文件预览后缀 |
| favicon_url    | string | 自定义favicon.ico      |
| footer         | string | 自定义底部信息         |
| head           | string | HEAD.md渲染            |
| image          | string | 图片文件预览后缀       |
| js             | string | 自定义javascript代码   |
| path_prefix    | string | 反代地址二级跳转       |
| readme         | string | README.md渲染          |
| s_column       | string | 默认排序字段           |
| s_order        | string | 默认排序方式           |
| short_action   | string | 短链行为               |
| site_name      | string | 网站名称               |
| theme          | string | 主题                   |
| video          | string | 视频文件预览后缀       |

#### 返回示例

```json
{
    "data": {
        "account_choose": "display",
        "audio": "mp3,wav,flac,ape",
        "code": "txt,go,html,js,java,json,css,lua,sh,sql,py,cpp,xml,jsp,properties,yaml,ini",
        "css": "",
        "doc": "doc,docx,dotx,ppt,pptx,xls,xlsx",
        "favicon_url": "",
        "footer": "",
        "head": "0",
        "image": "png,gif,jpg,bmp,jpeg,ico,webp",
        "js": "",
        "path_prefix": "",
        "readme": "1",
        "s_column": "file_name",
        "s_order": "asc",
        "short_action": "0",
        "site_name": "PanIndex",
        "theme": "mdui",
        "video": "mp4,mkv,m3u8,flv,avi"
    },
    "msg": "success",
    "status": 0
}
```

### 3. 获取目录、文件列表

| URL                  | Request Method | Content-Type        |
| -------------------- | -------------- | ------------------- |
| /api/v3/public/index | POST           | multipart/form-data |

#### 请求参数

| Name      | Type   | Description    |
| --------- | ------ | -------------- |
| path      | string | 文件、目录路径 |
| sort_by   | string | 排序字段       |
| order     | string | 排序方式       |
| page_no   | number | 页码           |
| page_size | number | 分页大小       |



#### 返回参数

| Name                    | Type    | Description                               |
| ----------------------- | ------- | ----------------------------------------- |
| is_folder               | boolean | 是否是目录                                |
| last_file               | string  | 上一个文件                                |
| next_file               | string  | 下一个文件                                |
| no_referrer             | boolean | 页面是否需要no_referrer，主要用于阿里云盘 |
| page_no                 | string  | 页码                                      |
| page_size               | number  | 分页大小                                  |
| pages                   | number  | 页码数                                    |
| total_count             | number  | 总条数                                    |
| content                 | object  | 列表                                      |
| content[0].file_name    | string  | 文件名称                                  |
| content[0].file_size    | long    | 文件大小                                  |
| content[0].size_fmt     | string  | 文件大小(人类可读)                        |
| content[0].file_type    | string  | 文件类型                                  |
| content[0].is_folder    | boolean | 是否是目录                                |
| content[0].last_op_time | string  | 最近一次操作时间                          |
| content[0].path         | string  | 文件路径                                  |
| content[0].thumbnail    | string  | 缩略图                                    |
| content[0].view_type    | string  | 预览类型                                  |

#### 返回示例

```json
{
    "data": {
        "content": [
            {
                "file_name": "jakob-owens-EwRM05V0VSI-unsplash.jpg",
                "file_size": 3539743,
                "size_fmt": "3.38 MB",
                "file_type": "jpg",
                "is_folder": false,
                "last_op_time": "2022-11-27 00:02:46",
                "path": "/native/密码(1234)/jakob-owens-EwRM05V0VSI-unsplash.jpg",
                "thumbnail": "",
                "download_url": "",
                "view_type": "img"
            }
        ],
        "is_folder": true,
        "last_file": "",
        "next_file": "",
        "no_referrer": false,
        "page_no": 1,
        "page_size": 10,
        "pages": 1
    },
    "msg": "success",
    "status": 0
}
```

### 4. 获取文件内容

| URL                       | Request Method | Content-Type                      |
| ------------------------- | -------------- | --------------------------------- |
| /api/v3/public/raw/{path} | GET            | application/x-www-form-urlencoded |

#### 请求参数

| Name | Type   | Description |
| ---- | ------ | ----------- |
| path | string | 文件路径    |

#### 返回参数

| Name | Type   | Description |
| ---- | ------ | ----------- |
| -    | string | 文件内容    |

#### 返回示例

```json
sss
```



### 5. 获取网盘列表

| URL                         | Request Method | Content-Type                      |
| --------------------------- | -------------- | --------------------------------- |
| /api/v3/public/account/list | GET            | application/x-www-form-urlencoded |

#### 请求参数

无

#### 返回参数

| Name | Type   | Description |
| ---- | ------ | ----------- |
| mode | string | 网盘类型    |
| name | string | 名称        |
| path | string | 路径        |

#### 返回示例

```json
{
    "data": [
        {
            "mode": "native",
            "name": "native",
            "path": "/native"
        },
    ],
    "msg": "success",
    "status": 0
}
```



### 6. 短链 & 二维码

| URL                      | Request Method | Content-Type        |
| ------------------------ | -------------- | ------------------- |
| /api/v3/public/shortInfo | POST           | multipart/form-data |

#### 请求参数

| Name   | Type    | Description                     |
| ------ | ------- |---------------------------------|
| prefix | string  | URL前缀：https://localhost:5238/s/ |
| path   | string  | 文件路径                            |
| isFile | boolean | 是否是文件                           |

#### 返回参数

| Name      | Type   | Description      |
| --------- | ------ | ---------------- |
| qr_code   | string | 二维码BASE64图片 |
| short_url | string | 短链             |

#### 返回示例

```json
{
    "msg": "短链生成成功",
    "qr_code": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAQAAAAEAAQMAAABmvDolAAAABlBMVEX///8AAABVwtN+AAABn0lEQVR42uyYMa7mIAyEB6VImSPkKBwNjsZRcoSUKSJmZRsidv/3t6uHFZfWV9nD2AZvvPHGr42NEhdiDrywF/DqKVfAAWCRZMVKKmApZ0BkXaQOWHjuBfVJeQSQsJAHEjwDW0nAegJJ2u0RMAlHyuNFzF9kPztghrSVFO713EsKP5vY5ECPSH285LfJMzewHTGbhHmvPGJWOod7MgCpSlNzEAAIZkSjqh0A2Mm6kMym06RGhEG0fgCsLFqNIzKI0w7t9gFsLMmctiqdqe0Og0d5AFodTiTaNov1B1XPDyDq/CjqUSyWG+vgAtjk2ZqEbaDYFjSq2gOAvv/IsidWbO3+u1AOgF3caWWbOHJ2Ugxr8QXIAisTp6Q+cTQCfQFSh3DL/tDOlSflCtA1SfcHye3Udis9E9C5Ig1UP7Jr698DZHLgOarUghRQp/UGPFdz1YGSw6XmO4rWBWB/ICLkNkq6xB0CsgvdbW6eiLnCJbC1Xaid1fz4hp0e6N9XrQ7iVAWfsp8caD9abeIU8JKj5cPEJgfeeOON/x5/AgAA//+RAbfF+WNWoQAAAABJRU5ErkJggg==",
    "short_url": "https://t2.noki.icu/s/3aiuAf"
}
```



### 7. 搜索

| URL                   | Request Method | Content-Type        |
| --------------------- | -------------- | ------------------- |
| /api/v3/public/search | POST           | multipart/form-data |

#### 请求参数

| Name | Type   | Description |
| ---- | ------ | ----------- |
| key  | string | 搜索关键字  |

#### 返回参数

> 同列表接口

#### 返回示例

```json
{
    "data": {
        "content": [
            {
                "file_name": "code_bash.sh",
                "file_size": 268,
                "size_fmt": "268.00 B",
                "file_type": "sh",
                "is_folder": false,
                "last_op_time": "2022-04-01 18:43:00",
                "path": "/cloud189/代码/code_bash.sh",
                "thumbnail": "",
                "download_url": "",
                "view_type": "code"
            }
        ]
    },
    "msg": "success",
    "status": 0
}
```



### 8. 短链跳转

| URL                  | Request Method | Content-Type        |
| -------------------- | -------------- | ------------------- |
| /api/v3/public/short | POST           | multipart/form-data |

#### 请求参数

| Name       | Type   | Description |
| ---------- | ------ | ----------- |
| short_code | string | 短链Code    |

#### 返回参数

| Name        | Type   | Description    |
| ----------- | ------ | -------------- |
| redirectUri | string | 重定向文件路径 |
| v           | string | 预览标识       |

#### 返回示例

```json
{
    "data": {
        "redirectUri": "/native/代码/code_bash.sh",
        "v": "v"
    },
    "msg": "success",
    "status": 0
}
```

### 9. 文件过滤

| URL                  | Request Method | Content-Type        |
| -------------------- | -------------- | ------------------- |
| /api/v3/public/files | POST           | multipart/form-data |

#### 请求参数

| Name       | Type   | Description |
| ---------- | ------ | ----------- |
| path       | string | 文件夹路径  |
| viewType   | string | 过滤类型    |
| sortColumn | string | 排序字段    |
| sortOrder  | string | 排序方式    |

#### 返回参数

> 同文件列表

#### 返回示例

```json
{
    "data": [
        {
            "file_name": "Goose house - 光るなら.mp3",
            "file_size": 10404056,
            "size_fmt": "9.92 MB",
            "file_type": "mp3",
            "is_folder": false,
            "last_op_time": "2022-11-27 00:03:17",
            "path": "/native/音频/Goose house - 光るなら.mp3",
            "thumbnail": "",
            "download_url": "",
            "view_type": "audio",
        }
    ],
    "message": "success",
    "status": 0
}
```




## 后台

待完善