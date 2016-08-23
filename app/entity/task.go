package entity

import (
	"bufio"
	"fmt"
	"strings"
	"time"
)

type Task struct {
	Id           int
	ProjectId    int       `orm:"index"`                       // 项目id
	StartVer     string    `orm:"size(20)"`                    // 起始版本号
	EndVer       string    `orm:"size(20)"`                    // 结束版本号
	Message      string    `orm:"type(text)"`                  // 版本说明
	UserId       int       `orm:"index"`                       // 创建人ID
	UserName     string    `orm:"size(20)"`                    // 创建人名称
	BuildStatus  int       `orm:"default(0)"`                  // 构建状态
	ChangeLogs   string    `orm:"type(text)"`                  // 修改日志列表
	ChangeFiles  string    `orm:"type(text)"`                  // 修改文件列表
	Filepath     string    `orm:"size(200)"`                   // 更新包路径
	PubEnvId     int       `orm:"default(0)"`                  // 发布环境ID
	PubStatus    int       `orm:"default(0)"`                  // 发布状态：1 正在发布，2 发布到跳板机，3 发布到目标服务器，-2 发布到跳板机失败，-3 发布到目标服务器失败
	PubTime      time.Time `orm:"null;type(datetime)"`         // 发布时间
	ErrorMsg     string    `orm:"type(text)"`                  // 错误消息
	PubLog       string    `orm:"type(text)"`                  // 发布日志
	ReviewStatus int       `orm:"default(0)"`                  // 审批状态
	CreateTime   time.Time `orm:"auto_now_add;type(datetime)"` // 创建时间
	UpdateTime   time.Time `orm:"auto_now;type(datetime)"`     // 更新时间
	ProjectInfo  *Project  `orm:"-"`                           // 项目信息
	EnvInfo      *Env      `orm:"-"`                           // 发布环境
}

func (t *Task) GetChangeFileStat() string {
	var modifyNum, addNum, deleteNum int
	scaner := bufio.NewScanner(strings.NewReader(t.ChangeFiles))
	for scaner.Scan() {
		line := scaner.Bytes()
		switch line[0] {
		case 'M':
			modifyNum++
		case 'A':
			addNum++
		case 'D':
			deleteNum++
		}
	}
	return fmt.Sprintf("总数：%d，新增：%d，修改：%d，删除：%d", modifyNum+addNum+deleteNum, addNum, modifyNum, deleteNum)
}

type TaskReview struct {
	Id         int
	TaskId     int       `orm:"default(0)"`                  // 任务id
	UserId     int       `orm:"default(0)"`                  // 审批人id
	UserName   string    `orm:"size(20)"`                    // 审批人
	Status     int       `orm:"default(0)"`                  // 审批结果(1:通过;0:不通过)
	Message    string    `orm:"type(text)"`                  // 审批说明
	CreateTime time.Time `orm:"auto_now_add;type(datetime)"` // 创建时间
}
