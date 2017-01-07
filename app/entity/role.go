package entity

import (
    "time"
)

// 角色
type Role struct {
    Id          int
    RoleName    string    // 角色名称
    ProjectIds  string    // 项目权限
    Description string    // 说明
    CreateTime  time.Time // 创建时间
    UpdateTime  time.Time // 更新时间
    PermList    []Perm    // 权限列表
    UserList    []User    // 用户列表
}

// 角色权限
type RolePerm struct {
    RoleId int    // 角色id
    Perm   string // 权限
}
