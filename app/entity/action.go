package entity

import (
	"time"
)

// 用户动作
type Action struct {
	Id         int
	Action     string    `orm:"size(20)"`                // 动作类型
	Actor      string    `orm:"size(20)"`                // 操作角色
	ObjectType string    `orm:"size(20)"`                // 操作对象类型
	ObjectId   int       `orm:"default(0)"`              // 操作对象id
	Extra      string    `orm:"size(1000)"`              // 额外信息
	CreateTime time.Time `orm:"auto_now;type(datetime)"` // 更新时间
	Message    string    `orm:"-"`                       // 格式化后的消息
}
