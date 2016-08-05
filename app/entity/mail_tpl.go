package entity

import (
	"time"
)

// 表结构
type MailTpl struct {
	Id         int
	UserId     int
	Name       string    `orm:"size(100)"`                   // 模板名
	Subject    string    `orm:"size(200)"`                   // 邮件主题
	Content    string    `orm:"type(text)"`                  // 邮件内容
	MailTo     string    `orm:"size(1000)"`                  // 预设收件人
	MailCc     string    `orm:"size(1000)"`                  // 预设抄送人
	CreateTime time.Time `orm:"auto_now_add;type(datetime)"` // 创建时间
	UpdateTime time.Time `orm:"auto_now;type(datetime)"`     // 更新时间
}
