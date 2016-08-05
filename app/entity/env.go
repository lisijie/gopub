package entity

import (
	"time"
)

// 发布环境
type Env struct {
	Id          int
	ProjectId   int       `orm:"index"`                       // 项目id
	Name        string    `orm:"size(20)"`                    // 发布环境名称
	SshUser     string    `orm:"size(20)"`                    // 发布帐号
	SshPort     string    `orm:"size(10)"`                    // SSH端口
	SshKey      string    `orm:"size(100)"`                   // SSH KEY路径
	PubDir      string    `orm:"size(100)"`                   // 发布目录
	BeforeShell string    `orm:"type(text)"`                  // 发布前执行的shell脚本
	AfterShell  string    `orm:"type(text)"`                  // 发布后执行的shell脚本
	ServerCount int       `orm:"default(0)"`                  // 服务器数量
	SendMail    int       `orm:"default(0)"`                  // 是否发送发版邮件通知
	MailTplId   int       `orm:"default(0)"`                  // 邮件模板id
	MailTo      string    `orm:"size(1000)"`                  // 邮件收件人
	MailCc      string    `orm:"size(1000)"`                  // 邮件抄送人
	CreateTime  time.Time `orm:"auto_now_add;type(datetime)"` // 创建时间
	UpdateTime  time.Time `orm:"auto_now;type(datetime)"`     // 更新时间
	ServerList  []Server  `orm:"-"`                           // 服务器列表
}

// 表结构
type EnvServer struct {
	Id        int
	ProjectId int `orm:"default(0)"`       // 项目id
	EnvId     int `orm:"default(0);index"` // 环境id
	ServerId  int `orm:"default(0)"`       // 服务器id
}
