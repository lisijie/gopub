package entity

import (
    "time"
)

type User struct {
    Id         int
    UserName   string           // 用户名
    Password   string           // 密码
    Salt       string           // 密码盐
    Sex        int              // 性别
    Email      string           // 邮箱
    LastLogin  time.Time        // 最后登录时间
    LastIp     string           // 最后登录IP
    Status     int              // 状态，0正常 -1禁用
    CreateTime time.Time        // 创建时间
    UpdateTime time.Time        // 更新时间
    RoleList   []Role `orm:"-"` // 角色列表
}

type UserRole struct {
    UserId int // 用户id
    RoleId int // 角色id
}
