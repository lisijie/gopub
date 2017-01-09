package controllers

import (
	"github.com/astaxie/beego"
	"github.com/lisijie/gopub/app/entity"
	"github.com/lisijie/gopub/app/service"
	"strings"
)

type RoleController struct {
	BaseController
}

func (c *RoleController) List() {
	roleList, err := service.RoleService.GetAllRoles()
	c.checkError(err)
	for k, role := range roleList {
		roleList[k].UserList, _ = service.UserService.GetUserListByRoleId(role.Id)
	}
	c.Data["pageTitle"] = "角色管理"
	c.Data["list"] = roleList
	c.display()
}

func (c *RoleController) Add() {
	if c.isPost() {
		role := &entity.Role{}
		role.RoleName = c.GetString("role_name")
		role.Description = c.GetString("description")
		if role.RoleName == "" {
			c.showMsg("角色名不能为空", MSG_ERR)
		}
		err := service.RoleService.AddRole(role)
		c.checkError(err)
		c.redirect(beego.URLFor("RoleController.List"))
	}
	c.Data["pageTitle"] = "创建角色"
	c.display()
}

func (c *RoleController) Edit() {
	id, _ := c.GetInt("id")
	role, err := service.RoleService.GetRole(id)
	c.checkError(err)

	if c.isPost() {
		role.RoleName = c.GetString("role_name")
		role.Description = c.GetString("description")
		err := service.RoleService.UpdateRole(role, "RoleName", "Description")
		c.checkError(err)
		c.redirect(beego.URLFor("RoleController.List"))
	}

	c.Data["pageTitle"] = "编辑角色"
	c.Data["role"] = role
	c.display()
}

func (c *RoleController) Del() {
	id, _ := c.GetInt("id")

	err := service.RoleService.DeleteRole(id)
	c.checkError(err)

	c.redirect(beego.URLFor("RoleController.List"))
}

func (c *RoleController) Perm() {
	id, _ := c.GetInt("id")
	role, err := service.RoleService.GetRole(id)
	c.checkError(err)

	if c.isPost() {
		pids := c.GetStrings("pids")
		perms := c.GetStrings("perms")
		if len(pids) == 0 {
			role.ProjectIds = ""
		} else {
			role.ProjectIds = strings.Join(pids, ",")
		}
		err := service.RoleService.UpdateRole(role, "ProjectIds")
		c.checkError(err)
		err = service.RoleService.SetPerm(role.Id, perms)
		c.checkError(err)
		c.redirect(beego.URLFor("RoleController.List"))
	}

	projectList, _ := service.ProjectService.GetAllProject()
	permList := service.SystemService.GetPermList()

	chkmap := make(map[string]string)
	for _, v := range role.PermList {
		chkmap[v.Key] = "checked"
	}
	if role.ProjectIds != "" {
		pids := strings.Split(role.ProjectIds, ",")
		for _, pid := range pids {
			chkmap[pid] = "checked"
		}
	}

	c.Data["pageTitle"] = "编辑权限"
	c.Data["permList"] = permList
	c.Data["projectList"] = projectList
	c.Data["role"] = role
	c.Data["chkmap"] = chkmap
	c.display()
}
