package entity

import (
	"time"
)

// 表结构
type Server struct {
	Id          int
	TypeId      int       // 0:普通服务器, 1:跳板机
	Ip          string    `orm:"size(20)"`  // 服务器IP
	Area        string    `orm:"size(20)"`  // 机房
	Description string    `orm:"size(200)"` // 服务器说明
	SshPort     int       // ssh端口
	SshUser     string    `orm:"size(50)"`                    // ssh用户
	SshPwd      string    `orm:"size(100)"`                   // ssh密码
	SshKey      string    `orm:"size(100)"`                   // ssh key路径
	WorkDir     string    `orm:"size(100)"`                   // 工作目录
	CreateTime  time.Time `orm:"auto_now_add;type(datetime)"` // 创建时间
	UpdateTime  time.Time `orm:"auto_now;type(datetime)"`     // 更新时间
}
