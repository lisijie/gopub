// Copyright 2015 lisijie. All Rights Reserved.

package controllers

import (
    "fmt"
    "github.com/astaxie/beego"
    "github.com/astaxie/beego/validation"
    "github.com/lisijie/gopub/app/libs/utils"
    "github.com/lisijie/gopub/app/service"
    "regexp"
    "strconv"
)

type UserController struct {
    BaseController
}

// 帐号列表
func (c *UserController) List() {
    page, _ := strconv.Atoi(c.GetString("page"))
    if page < 1 {
        page = 1
    }

    count, _ := service.UserService.GetTotal()
    users, _ := service.UserService.GetUserList(page, c.pageSize, true)

    c.Data["pageTitle"] = "帐号管理"
    c.Data["count"] = count
    c.Data["list"] = users
    c.Data["pageBar"] = utils.NewPager(page, int(count), c.pageSize, beego.URLFor("UserController.List"), true).ToString()
    c.display()
}

// 添加帐号
func (c *UserController) Add() {
    if c.isPost() {
        valid := validation.Validation{}

        username := c.GetString("username")
        email := c.GetString("email")
        sex, _ := c.GetInt("sex")
        password1 := c.GetString("password1")
        password2 := c.GetString("password2")

        valid.Required(username, "username").Message("请输入用户名")
        valid.Required(email, "email").Message("请输入Email")
        valid.Email(email, "email").Message("Email无效")
        valid.Required(password1, "password1").Message("请输入密码")
        valid.Required(password2, "password2").Message("请输入确认密码")
        valid.MinSize(password1, 6, "password1").Message("密码长度不能小于6个字符")
        valid.Match(password1, regexp.MustCompile(`^` + regexp.QuoteMeta(password2) + `$`), "password2").Message("两次输入的密码不一致")
        if valid.HasErrors() {
            for _, err := range valid.Errors {
                c.showMsg(err.Message, MSG_ERR)
            }
        }

        user, err := service.UserService.AddUser(username, email, password1, sex)
        c.checkError(err)

        // 更新角色
        roleIds := make([]int, 0)
        for _, v := range c.GetStrings("role_ids") {
            if roleId, _ := strconv.Atoi(v); roleId > 0 {
                roleIds = append(roleIds, roleId)
            }
        }
        service.UserService.UpdateUserRoles(user.Id, roleIds)

        c.redirect(beego.URLFor("UserController.List"))
    }

    roleList, _ := service.RoleService.GetAllRoles()
    c.Data["pageTitle"] = "添加帐号"
    c.Data["roleList"] = roleList
    c.display()
}

func (c *UserController) Edit() {
    id, _ := c.GetInt("id")
    user, err := service.UserService.GetUser(id, true)
    c.checkError(err)

    if c.isPost() {
        valid := validation.Validation{}

        email := c.GetString("email")
        sex, _ := c.GetInt("sex")
        status, _ := c.GetInt("status")
        password1 := c.GetString("password1")
        password2 := c.GetString("password2")

        valid.Required(email, "email").Message("请输入Email")
        valid.Email(email, "email").Message("Email无效")
        if password1 != "" {
            valid.Required(password1, "password1").Message("请输入密码")
            valid.Required(password2, "password2").Message("请输入确认密码")
            valid.MinSize(password1, 6, "password1").Message("密码长度不能小于6个字符")
            valid.Match(password1, regexp.MustCompile(`^` + regexp.QuoteMeta(password2) + `$`), "password2").Message("两次输入的密码不一致")
        }

        if valid.HasErrors() {
            for _, err := range valid.Errors {
                c.showMsg(err.Message, MSG_ERR)
            }
        }

        user.Sex = sex
        user.Status = status
        user.Email = email
        service.UserService.UpdateUser(user, "Sex", "Status", "Email")

        if password1 != "" {
            service.UserService.ModifyPassword(user.Id, password1)
        }

        // 更新角色
        roleIds := make([]int, 0)
        for _, v := range c.GetStrings("role_ids") {
            if roleId, _ := strconv.Atoi(v); roleId > 0 {
                roleIds = append(roleIds, roleId)
            }
        }
        service.UserService.UpdateUserRoles(user.Id, roleIds)

        c.redirect(beego.URLFor("UserController.List"))
    }

    chkmap := make(map[int]string)
    for _, v := range user.RoleList {
        chkmap[v.Id] = "selected"
    }

    roleList, _ := service.RoleService.GetAllRoles()
    c.Data["pageTitle"] = "修改帐号"
    c.Data["user"] = user
    c.Data["roleList"] = roleList
    c.Data["chkmap"] = chkmap
    c.Data[fmt.Sprintf("sex%d", user.Sex)] = "checked"
    c.display()
}

func (c *UserController) Del() {
    id, _ := c.GetInt("id")

    if id == 1 {
        c.showMsg("不能删除ID为1的帐号。", MSG_ERR)
    }

    err := service.UserService.DeleteUser(id)
    c.checkError(err)

    c.redirect(beego.URLFor("UserController.List"))
}
