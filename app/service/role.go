package service

import (
	"errors"
	"../entity"
)

type roleService struct{}

func (this *roleService) table() string {
	return tableName("role")
}

// 根据id获取角色信息
func (this *roleService) GetRole(id int) (*entity.Role, error) {
	role := &entity.Role{
		Id: id,
	}
	err := o.Read(role)
	if err != nil {
		return nil, err
	}
	this.loadRoleExtra(role)
	return role, err
}

// 根据名称获取角色
func (this *roleService) GetRoleByName(roleName string) (*entity.Role, error) {
	role := &entity.Role{
		RoleName: roleName,
	}
	if err := o.Read(role, "RoleName"); err != nil {
		return nil, err
	}
	this.loadRoleExtra(role)
	return role, nil
}

func (this *roleService) loadRoleExtra(role *entity.Role) {
	o.Raw("SELECT SUBSTRING_INDEX(perm, '.', 1) as module,SUBSTRING_INDEX(perm, '.', -1) as `action`, perm AS `key` FROM "+tableName("role_perm")+" WHERE role_id = ?", role.Id).QueryRows(&role.PermList)
}

// 添加角色
func (this *roleService) AddRole(role *entity.Role) error {
	if _, err := this.GetRoleByName(role.RoleName); err == nil {
		return errors.New("角色已存在")
	}
	_, err := o.Insert(role)
	return err
}

// 获取所有角色列表
func (this *roleService) GetAllRoles() ([]entity.Role, error) {
	var (
		roles []entity.Role // 角色列表
	)
	if _, err := o.QueryTable(this.table()).All(&roles); err != nil {
		return nil, err
	}
	return roles, nil
}

// 更新角色信息
func (this *roleService) UpdateRole(role *entity.Role, fields ...string) error {
	if v, err := this.GetRoleByName(role.RoleName); err == nil && v.Id != role.Id {
		return errors.New("角色名称已存在")
	}
	_, err := o.Update(role, fields...)
	return err
}

// 设置角色权限
func (this *roleService) SetPerm(roleId int, perms []string) error {
	if _, err := this.GetRole(roleId); err != nil {
		return err
	}
	all := SystemService.GetPermList()
	pmmap := make(map[string]bool)
	for _, list := range all {
		for _, perm := range list {
			pmmap[perm.Key] = true
		}
	}
	for _, v := range perms {
		if _, ok := pmmap[v]; !ok {
			return errors.New("权限名称无效:" + v)
		}
	}
	o.Raw("DELETE FROM "+tableName("role_perm")+" WHERE role_id = ?", roleId).Exec()
	for _, v := range perms {
		o.Raw("REPLACE INTO "+tableName("role_perm")+" (role_id, perm) VALUES (?, ?)", roleId, v).Exec()
	}
	return nil
}

// 删除角色
func (this *roleService) DeleteRole(id int) error {
	role, err := this.GetRole(id)
	if err != nil {
		return err
	}
	o.Delete(role)
	o.Raw("DELETE FROM "+tableName("role_user")+" WHERE role_id = ?", id).Exec()
	return nil
}
