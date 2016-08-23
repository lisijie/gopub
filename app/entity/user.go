package entity

import (
	"time"
)

type User struct {
	Id         int
	UserName   string    `orm:"unique;size(20)"`             // 用户名
	Password   string    `orm:"size(32)"`                    // 密码
	Salt       string    `orm:"size(10)"`                    // 密码盐
	Sex        int       `orm:"default(0)"`                  // 性别
	Email      string    `orm:"size(50)"`                    // 邮箱
	LastLogin  time.Time `orm:"null;type(datetime)"`         // 最后登录时间
	LastIp     string    `orm:"size(15)"`                    // 最后登录IP
	Status     int       `orm:"default(0)"`                  // 状态，0正常 -1禁用
	CreateTime time.Time `orm:"auto_now_add;type(datetime)"` // 创建时间
	UpdateTime time.Time `orm:"auto_now;type(datetime)"`     // 更新时间
	RoleList   []Role    `orm:"-"`                           // 角色列表
}

type UserRole struct {
	UserId int // 用户id
	RoleId int // 角色id
}
