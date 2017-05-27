# 阿里云日志服务loghub GO语言封装

##相关功能
* 实现了阿里云请求签名
* 实现了protocol buffer封装
* 请求通过zlib压缩，节省上行流量
* 目前仅实现推送日至到日志服务

##相关用法

### 获取扩展包
```
go get github.com/zjlletian/ali_loghub
```

### 相关示例
```
import (
    "github.com/golang/protobuf/proto"
    "github.com/zjlletian/ali_loghub/loghub"
    "time"
)
...

//loghub配置
conf := loghub.Config{
    AccessKey:    "key id",
    AccessSecret: "key secret",
    EndPoint:     "{project_name}.{region}.log.aliyuncs.com",
    LogStore:     "{logstore_name}",
}

//构造日志，单个日志由多个kv构成
kv1 := &loghub.Log_Content{
    Key:   proto.String("k1"),
    Value: proto.String("v1"),
}
kv2 := &loghub.Log_Content{
    Key:   proto.String("k2"),
    Value: proto.String("v2"),
}
contents := []*loghub.Log_Content{kv1, kv2}
now := uint32(time.Now().Unix())
log := &loghub.Log{
    Time:     &now,
    Contents: contents,
}

//每次可发送多个日志，最大不超过4096条（依据阿里云限制）
logs := []*loghub.Log{log, log, log, log}
err := loghub.SendLog(conf, logs)
if err !=nil {
    ....
}

```