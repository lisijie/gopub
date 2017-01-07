package entity

import (
    "time"
)

// 表结构
type MailTpl struct {
    Id         int
    UserId     int
    Name       string    // 模板名
    Subject    string    // 邮件主题
    Content    string    // 邮件内容
    MailTo     string    // 预设收件人
    MailCc     string    // 预设抄送人
    CreateTime time.Time // 创建时间
    UpdateTime time.Time // 更新时间
}
