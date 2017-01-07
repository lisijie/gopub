package entity

import (
    "time"
)

// 表结构
type Log struct {
    Id         int
    UserId     int
    Action     string    // 操作
    Url        string    // url
    PostData   string    // 提交数据
    Message    string    // 消息内容
    CreateTime time.Time // 创建时间
}
