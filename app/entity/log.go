package entity

import (
	"time"
)

// 表结构
type Log struct {
	Id         int
	UserId     int
	Action     string    `orm:"size(100)"`                   // 操作
	Url        string    `orm:"size(200)"`                   // url
	PostData   string    `orm:"type(text)"`                  // 提交数据
	Message    string    `orm:"type(text)"`                  // 消息内容
	CreateTime time.Time `orm:"auto_now_add;type(datetime)"` // 创建时间
}
