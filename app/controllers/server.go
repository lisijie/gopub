package controllers

import (
    "github.com/astaxie/beego"
    "github.com/astaxie/beego/validation"
    "github.com/lisijie/gopub/app/entity"
    "github.com/lisijie/gopub/app/libs/utils"
    "github.com/lisijie/gopub/app/service"
    "strconv"
)

type ServerController struct {
    BaseController
}

// 列表
func (c *ServerController) List() {
    page, _ := strconv.Atoi(c.GetString("page"))
    if page < 1 {
        page = 1
    }
    count, err := service.ServerService.GetTotal(service.SERVER_TYPE_NORMAL)
    c.checkError(err)
    serverList, err := service.ServerService.GetServerList(page, c.pageSize)
    c.checkError(err)

    c.Data["count"] = count
    c.Data["list"] = serverList
    c.Data["pageBar"] = utils.NewPager(page, int(count), c.pageSize, beego.URLFor("ServerController.List"), true).ToString()
    c.Data["pageTitle"] = "服务器列表"
    c.display()
}

// 添加
func (c *ServerController) Add() {
    if c.isPost() {
        valid := validation.Validation{}
        server := &entity.Server{}
        server.TypeId = service.SERVER_TYPE_NORMAL
        server.Ip = c.GetString("server_ip")
        server.Area = c.GetString("area")
        server.Description = c.GetString("description")
        valid.Required(server.Ip, "ip").Message("请输入服务器IP")
        valid.IP(server.Ip, "ip").Message("服务器IP无效")
        if valid.HasErrors() {
            for _, err := range valid.Errors {
                c.showMsg(err.Message, MSG_ERR)
            }
        }

        if err := service.ServerService.AddServer(server); err != nil {
            c.showMsg(err.Error(), MSG_ERR)
        }

        c.redirect(beego.URLFor("ServerController.List"))
    }

    c.Data["pageTitle"] = "添加服务器"
    c.display()
}

// 编辑
func (c *ServerController) Edit() {
    id, _ := c.GetInt("id")
    server, err := service.ServerService.GetServer(id, service.SERVER_TYPE_NORMAL)
    c.checkError(err)

    if c.isPost() {
        valid := validation.Validation{}
        ip := c.GetString("server_ip")
        server.Area = c.GetString("area")
        server.Description = c.GetString("description")
        valid.Required(ip, "ip").Message("请输入服务器IP")
        valid.IP(ip, "ip").Message("服务器IP无效")
        if valid.HasErrors() {
            for _, err := range valid.Errors {
                c.showMsg(err.Message, MSG_ERR)
            }
        }
        server.Ip = ip
        err := service.ServerService.UpdateServer(server)
        c.checkError(err)
        c.redirect(beego.URLFor("ServerController.List"))
    }

    c.Data["pageTitle"] = "编辑服务器"
    c.Data["server"] = server
    c.display()
}

// 删除
func (c *ServerController) Del() {
    id, _ := c.GetInt("id")

    _, err := service.ServerService.GetServer(id, service.SERVER_TYPE_NORMAL)
    c.checkError(err)

    err = service.ServerService.DeleteServer(id)
    c.checkError(err)

    c.redirect(beego.URLFor("ServerController.List"))
}

// 项目列表
func (c *ServerController) Projects() {
    id, _ := c.GetInt("id")
    server, err := service.ServerService.GetServer(id, service.SERVER_TYPE_NORMAL)
    c.checkError(err)
    envList, err := service.EnvService.GetEnvListByServerId(id)
    c.checkError(err)

    result := make(map[int]map[string]interface{})
    for _, env := range envList {
        if _, ok := result[env.ProjectId]; !ok {
            project, err := service.ProjectService.GetProject(env.ProjectId)
            if err != nil {
                continue
            }
            row := make(map[string]interface{})
            row["projectId"] = project.Id
            row["projectName"] = project.Name
            row["envName"] = env.Name
            result[env.ProjectId] = row
        } else {
            result[env.ProjectId]["envName"] = result[env.ProjectId]["envName"].(string) + ", " + env.Name
        }
    }

    c.Data["list"] = result
    c.Data["server"] = server
    c.Data["pageTitle"] = server.Ip + " 下的项目列表"
    c.display()
}
