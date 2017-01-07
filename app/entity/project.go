package entity

import (
    "time"
)

// 表结构
type Project struct {
    Id            int
    Name          string    // 项目名称
    Domain        string    // 项目标识
    Version       string    // 最后发布版本
    VersionTime   time.Time // 最后发版时间
    RepoType      string    // 仓库类型（GIT|SVN）
    RepoUrl       string    // 仓库地址
    RepoUser      string    // 仓库用户
    RepoPassword  string    // 仓库密码
    Status        int       // 初始化状态
    ErrorMsg      string    // 错误消息
    AgentId       int       // 跳板机ID
    IgnoreList    string    // 忽略文件列表
    BeforeShell   string    // 发布前要执行的shell脚本
    AfterShell    string    // 发布后要执行的shell脚本
    CreateVerfile int       // 是否生成版本号文件
    VerfilePath   string    // 版本号文件目录
    TaskReview    int       // 发布单是否需要经过审批
    CreateTime    time.Time // 创建时间
    UpdateTime    time.Time // 更新时间
}
