package entity

import (
    "bufio"
    "fmt"
    "strings"
    "time"
)

type Task struct {
    Id           int
    ProjectId    int                // 项目id
    StartVer     string             // 起始版本号
    EndVer       string             // 结束版本号
    Message      string             // 版本说明
    UserId       int                // 创建人ID
    UserName     string             // 创建人名称
    BuildStatus  int                // 构建状态
    ChangeLogs   string             // 修改日志列表
    ChangeFiles  string             // 修改文件列表
    Filepath     string             // 更新包路径
    PubEnvId     int                // 发布环境ID
    PubStatus    int                // 发布状态：1 正在发布，2 发布到跳板机，3 发布到目标服务器，-2 发布到跳板机失败，-3 发布到目标服务器失败
    PubTime      time.Time          // 发布时间
    ErrorMsg     string             // 错误消息
    PubLog       string             // 发布日志
    ReviewStatus int                // 审批状态
    CreateTime   time.Time          // 创建时间
    UpdateTime   time.Time          // 更新时间
    ProjectInfo  *Project `orm:"-"` // 项目信息
    EnvInfo      *Env `orm:"-"`     // 发布环境
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
    return fmt.Sprintf("总数：%d，新增：%d，修改：%d，删除：%d", modifyNum + addNum + deleteNum, addNum, modifyNum, deleteNum)
}

type TaskReview struct {
    Id         int
    TaskId     int       // 任务id
    UserId     int       // 审批人id
    UserName   string    // 审批人
    Status     int       // 审批结果(1:通过;0:不通过)
    Message    string    // 审批说明
    CreateTime time.Time // 创建时间
}
