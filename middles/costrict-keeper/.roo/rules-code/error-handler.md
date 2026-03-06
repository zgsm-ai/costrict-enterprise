# RESTful API错误处理

RESTful API发生错误，统一返回如下格式信息：

```json
{
    "code": "group.errortag",
    "error": "detail information"
}
```

其中code表示错误类型，用于程序分别错误类别，message提供详细信息供人类阅读。

