package entity

import (
	"time"
)

// 表结构
type Project struct {
	Id            int
	Name          string    `orm:"size(100)"`                   // 项目名称
	Domain        string    `orm:"size(100)"`                   // 项目标识
	Version       string    `orm:"size(20)"`                    // 最后发布版本
	VersionTime   time.Time `orm:"type(datetime)"`              // 最后发版时间
	RepoUrl       string    `orm:"size(100)"`                   // 仓库地址
	Status        int       `orm:"default(0)"`                  // 初始化状态
	ErrorMsg      string    `orm:"type(text)"`                  // 错误消息
	AgentId       int       `orm:"default(0)"`                  // 跳板机ID
	IgnoreList    string    `orm:"type(text)"`                  // 忽略文件列表
	BeforeShell   string    `orm:"type(text)"`                  // 发布前要执行的shell脚本
	AfterShell    string    `orm:"type(text)"`                  // 发布后要执行的shell脚本
	CreateVerfile int       `orm:"default(0)"`                  // 是否生成版本号文件
	VerfilePath   string    `orm:"size(50)"`                    // 版本号文件目录
	TaskReview    int       `orm:"default(0)"`                  // 发布单是否需要经过审批
	CreateTime    time.Time `orm:"auto_now_add;type(datetime)"` // 创建时间
	UpdateTime    time.Time `orm:"auto_now;type(datetime)"`     // 更新时间
}
