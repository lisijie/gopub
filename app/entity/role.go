package entity

import (
    "time"
)

// 角色
type Role struct {
    Id          int
    Name    string           // 角色名称
    ProjectIds  string           // 项目权限
    Description string           // 说明
    CreateTime  time.Time        // 创建时间
    UpdateTime  time.Time        // 更新时间
    PermList    []Perm `orm:"-"` // 权限列表
    UserList    []User `orm:"-"` // 用户列表
}

// 角色权限
type RolePerm struct {
    RoleId int    // 角色id
    Perm   string // 权限
}

