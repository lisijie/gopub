package service

import (
    "errors"
    "fmt"
    "github.com/astaxie/beego"
    "github.com/astaxie/beego/orm"
    "github.com/lisijie/gopub/app/entity"
    "github.com/lisijie/gopub/app/libs/utils"
    "strconv"
    "strings"
    "time"
)

// 登录验证服务
// 提供登录验证、权限检查接口
type AuthService struct {
    loginUser *entity.User    // 当前登录用户
    permMap   map[string]bool // 当前用户权限表
    openPerm  map[string]bool // 公开的权限
}

func NewAuth(token string) *AuthService {
    s := &AuthService{}
    s.init(token)
    return s
}

// 初始化开放权限
func (s *AuthService) initOpenPerm() {
    s.openPerm = map[string]bool{
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
func (s *AuthService) GetUser() *entity.User {
    return s.loginUser
}

// 获取当前登录的用户id
func (s *AuthService) GetUserId() int {
    if s.IsLogined() {
        return s.loginUser.Id
    }
    return 0
}

// 获取当前登录的用户名
func (s *AuthService) GetUserName() string {
    if s.IsLogined() {
        return s.loginUser.UserName
    }
    return ""
}

// 初始化
func (s *AuthService) init(token string) {
    s.initOpenPerm()
    arr := strings.Split(token, "|")
    beego.Trace("登录验证, token: ", token)
    if len(arr) == 2 {
        idStr, password := arr[0], arr[1]
        userId, _ := strconv.Atoi(idStr)
        if userId > 0 {
            user, err := UserService.GetUser(userId, true)
            if err == nil && password == utils.Md5([]byte(user.Password + user.Salt)) {
                s.loginUser = user
                s.initPermMap()
            }
        }
    }
}

// 初始化权限表
func (s *AuthService) initPermMap() {
    s.permMap = make(map[string]bool)
    for _, role := range s.loginUser.RoleList {
        for _, perm := range role.PermList {
            s.permMap[perm.Key] = true
        }
    }
}

// 检查是否有某个权限
func (s *AuthService) HasAccessPerm(module, action string) bool {
    key := module + "." + action
    if !s.IsLogined() {
        return false
    }
    if s.loginUser.Id == 1 || s.isOpenPerm(key) {
        return true
    }
    if _, ok := s.permMap[key]; ok {
        return true
    }
    return false
}

// 检查是否登录
func (s *AuthService) IsLogined() bool {
    return s.loginUser != nil && s.loginUser.Id > 0
}

// 是否公开访问的操作
func (s *AuthService) isOpenPerm(key string) bool {
    if _, ok := s.openPerm[key]; ok {
        return true
    }
    return false
}

// 用户登录
func (s *AuthService) Login(userName, password string) (string, error) {
    user, err := UserService.GetUserByName(userName)
    if err != nil {
        if err == orm.ErrNoRows {
            return "", errors.New("帐号或密码错误")
        } else {
            return "", errors.New("系统错误")
        }
    }

    if user.Password != utils.Md5([]byte(password + user.Salt)) {
        return "", errors.New("帐号或密码错误")
    }
    if user.Status == -1 {
        return "", errors.New("该帐号已禁用")
    }

    user.LastLogin = time.Now()
    UserService.UpdateUser(user, "LastLogin")
    s.loginUser = user

    token := fmt.Sprintf("%d|%s", user.Id, utils.Md5([]byte(user.Password + user.Salt)))
    return token, nil
}

// 退出登录
func (s *AuthService) Logout() error {
    return nil
}
