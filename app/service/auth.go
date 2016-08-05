package service

import (
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/lisijie/gopub/app/entity"
	"github.com/lisijie/gopub/app/libs"
	"strconv"
	"strings"
	"time"
)

// 登录验证服务
type AuthService struct {
	loginUser *entity.User    // 当前登录用户
	permMap   map[string]bool // 当前用户权限表
	openPerm  map[string]bool // 公开的权限
}

func NewAuth() *AuthService {
	return new(AuthService)
}

// 初始化开放权限
func (this *AuthService) initOpenPerm() {
	this.openPerm = map[string]bool{
		"main.index":        true,
		"main.profile":      true,
		"main.login":        true,
		"main.logout":       true,
		"main.getpubstat":   true,
		"project.clone":     true,
		"project.getstatus": true,
		"task.gettags":      true,
		"task.getstatus":    true,
		"task.startpub":     true,
	}
}

// 获取当前登录的用户对象
func (this *AuthService) GetUser() *entity.User {
	return this.loginUser
}

// 获取当前登录的用户id
func (this *AuthService) GetUserId() int {
	if this.IsLogined() {
		return this.loginUser.Id
	}
	return 0
}

// 获取当前登录的用户名
func (this *AuthService) GetUserName() string {
	if this.IsLogined() {
		return this.loginUser.UserName
	}
	return ""
}

// 初始化
func (this *AuthService) Init(token string) {
	this.initOpenPerm()
	arr := strings.Split(token, "|")
	beego.Trace("登录验证, token: ", token)
	if len(arr) == 2 {
		idstr, password := arr[0], arr[1]
		userId, _ := strconv.Atoi(idstr)
		if userId > 0 {
			user, err := UserService.GetUser(userId, true)
			if err == nil && password == libs.Md5([]byte(user.Password+user.Salt)) {
				this.loginUser = user
				this.initPermMap()
				beego.Trace("验证成功，用户信息: ", user)
			}
		}
	}
}

// 初始化权限表
func (this *AuthService) initPermMap() {
	this.permMap = make(map[string]bool)
	for _, role := range this.loginUser.RoleList {
		for _, perm := range role.PermList {
			this.permMap[perm.Key] = true
		}
	}
}

// 检查是否有某个权限
func (this *AuthService) HasAccessPerm(module, action string) bool {
	key := module + "." + action
	if !this.IsLogined() {
		return false
	}
	if this.loginUser.Id == 1 || this.isOpenPerm(key) {
		return true
	}
	if _, ok := this.permMap[key]; ok {
		return true
	}
	return false
}

// 检查是否登录
func (this *AuthService) IsLogined() bool {
	return this.loginUser != nil && this.loginUser.Id > 0
}

// 是否公开访问的操作
func (this *AuthService) isOpenPerm(key string) bool {
	if _, ok := this.openPerm[key]; ok {
		return true
	}
	return false
}

// 用户登录
func (this *AuthService) Login(userName, password string) (string, error) {
	user, err := UserService.GetUserByName(userName)
	if err != nil {
		if err == orm.ErrNoRows {
			return "", errors.New("帐号或密码错误")
		} else {
			return "", errors.New("系统错误")
		}
	}

	if user.Password != libs.Md5([]byte(password+user.Salt)) {
		return "", errors.New("帐号或密码错误")
	}
	if user.Status == -1 {
		return "", errors.New("该帐号已禁用")
	}

	user.LastLogin = time.Now()
	UserService.UpdateUser(user, "LastLogin")
	this.loginUser = user

	token := fmt.Sprintf("%d|%s", user.Id, libs.Md5([]byte(user.Password+user.Salt)))
	return token, nil
}

// 退出登录
func (this *AuthService) Logout() error {
	return nil
}
