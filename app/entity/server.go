package entity

import (
    "time"
)

// 表结构
type Server struct {
    Id          int
    TypeId      int       // 0:普通服务器, 1:跳板机
    Ip          string    // 服务器IP
    Area        string    // 机房
    Description string    // 服务器说明
    SshPort     int       // ssh端口
    SshUser     string    // ssh用户
    SshPwd      string    // ssh密码
    SshKey      string    // ssh key路径
    WorkDir     string    // 工作目录
    CreateTime  time.Time // 创建时间
    UpdateTime  time.Time // 更新时间
}
