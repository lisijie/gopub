package entity

import (
    "time"
)

// 用户动作
type Action struct {
    Id         int
    Action     string           // 动作类型
    Actor      string           // 操作角色
    ObjectType string           // 操作对象类型
    ObjectId   int              // 操作对象id
    Extra      string           // 额外信息
    CreateTime time.Time        // 更新时间
    Message    string `orm:"-"` // 格式化后的消息
}
