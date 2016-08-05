package service

import (
	"errors"
	"github.com/astaxie/beego/utils"
	"github.com/lisijie/gopub/app/entity"
	"github.com/lisijie/gopub/app/libs"
	"time"
)

type userService struct{}

func (this *userService) table() string {
	return tableName("user")
}

// 根据用户id获取一个用户信息
func (this *userService) GetUser(userId int, getRoleInfo bool) (*entity.User, error) {
	user := &entity.User{}
	user.Id = userId

	err := o.Read(user)
	if err == nil && getRoleInfo {
		user.RoleList, _ = this.GetUserRoleList(user.Id)
	}
	return user, err
}

// 根据用户名获取用户信息
func (this *userService) GetUserByName(userName string) (*entity.User, error) {
	user := &entity.User{}
	user.UserName = userName
	err := o.Read(user, "UserName")
	return user, err
}

// 获取用户总数
func (this *userService) GetTotal() (int64, error) {
	return o.QueryTable(this.table()).Count()
}

// 分页获取用户列表
func (this *userService) GetUserList(page, pageSize int, getRoleInfo bool) ([]entity.User, error) {
	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}

	var users []entity.User
	qs := o.QueryTable(this.table())
	_, err := qs.OrderBy("id").Limit(pageSize, offset).All(&users)
	for k, user := range users {
		users[k].RoleList, _ = this.GetUserRoleList(user.Id)
	}

	return users, err
}

// 根据角色id获取用户列表
func (this *userService) GetUserListByRoleId(roleId int) ([]entity.User, error) {
	var users []entity.User
	sql := "SELECT u.* FROM " + this.table() + " u JOIN " + tableName("user_role") + " r ON u.id = r.user_id WHERE r.role_id = ?"
	_, err := o.Raw(sql, roleId).QueryRows(&users)
	return users, err
}

// 获取某个用户的角色列表
// 为什么不直接连表查询role表？因为不想“越权”查询
func (this *userService) GetUserRoleList(userId int) ([]entity.Role, error) {
	var (
		roleRef  []entity.UserRole
		roleList []entity.Role
	)
	sql := "SELECT role_id FROM " + tableName("user_role") + " WHERE user_id = ?"
	o.Raw(sql, userId).QueryRows(&roleRef)

	roleList = make([]entity.Role, 0, len(roleRef))
	for _, v := range roleRef {
		if role, err := RoleService.GetRole(v.RoleId); err == nil {
			roleList = append(roleList, *role)
		}
	}
	return roleList, nil
}

// 添加用户
func (this *userService) AddUser(userName, email, password string, sex int) (*entity.User, error) {
	if exists, _ := this.GetUserByName(userName); exists.Id > 0 {
		return nil, errors.New("用户名已存在")
	}

	user := &entity.User{}
	user.UserName = userName
	user.Sex = sex
	user.Email = email
	user.Salt = string(utils.RandomCreateBytes(10))
	user.Password = libs.Md5([]byte(password + user.Salt))
	user.LastLogin = time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)
	_, err := o.Insert(user)
	return user, err
}

// 更新用户信息
func (this *userService) UpdateUser(user *entity.User, fileds ...string) error {
	if len(fileds) < 1 {
		return errors.New("更新字段不能为空")
	}
	_, err := o.Update(user, fileds...)
	return err
}

// 修改密码
func (this *userService) ModifyPassword(userId int, password string) error {
	user, err := this.GetUser(userId, false)
	if err != nil {
		return err
	}
	user.Salt = string(utils.RandomCreateBytes(10))
	user.Password = libs.Md5([]byte(password + user.Salt))
	_, err = o.Update(user, "Salt", "Password")
	return err
}

// 删除用户
func (this *userService) DeleteUser(userId int) error {
	if userId == 1 {
		return errors.New("不允许删除用户ID为1的用户")
	}
	user := &entity.User{
		Id: userId,
	}
	_, err := o.Delete(user)
	return err
}

// 设置用户角色
func (this *userService) UpdateUserRoles(userId int, roleIds []int) error {
	if _, err := this.GetUser(userId, false); err != nil {
		return err
	}
	o.Raw("DELETE FROM "+tableName("user_role")+" WHERE user_id = ?", userId).Exec()
	for _, v := range roleIds {
		o.Raw("INSERT INTO "+tableName("user_role")+" (user_id, role_id) VALUES (?, ?)", userId, v).Exec()
	}
	return nil
}
