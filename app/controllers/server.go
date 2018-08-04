package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/astaxie/beego/validation"
	"github.com/lisijie/gopub/app/entity"
	"github.com/lisijie/gopub/app/libs"
	"github.com/lisijie/gopub/app/service"
	"strconv"
)

type ServerController struct {
	BaseController
}

// 列表
func (this *ServerController) List() {
	page, _ := strconv.Atoi(this.GetString("page"))
	if page < 1 {
		page = 1
	}
	count, err := service.ServerService.GetTotal(service.SERVER_TYPE_NORMAL)
	if err != orm.ErrNoRows {
		this.checkError(err)
	}
	serverList, err := service.ServerService.GetServerList(page, this.pageSize)
	if err != orm.ErrNoRows {
		this.checkError(err)
	}

	this.Data["count"] = count
	this.Data["list"] = serverList
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("ServerController.List"), true).ToString()
	this.Data["pageTitle"] = "服务器列表"
	this.display()
}

// 添加
func (this *ServerController) Add() {
	if this.isPost() {
		valid := validation.Validation{}
		server := &entity.Server{}
		server.TypeId = service.SERVER_TYPE_NORMAL
		server.Ip = this.GetString("server_ip")
		server.Area = this.GetString("area")
		server.Description = this.GetString("description")
		valid.Required(server.Ip, "ip").Message("请输入服务器IP")
		valid.IP(server.Ip, "ip").Message("服务器IP无效")
		if valid.HasErrors() {
			for _, err := range valid.Errors {
				this.showMsg(err.Message, MSG_ERR)
			}
		}

		if err := service.ServerService.AddServer(server); err != nil {
			this.showMsg(err.Error(), MSG_ERR)
		}

		this.redirect(beego.URLFor("ServerController.List"))
	}

	this.Data["pageTitle"] = "添加服务器"
	this.display()
}

// 编辑
func (this *ServerController) Edit() {
	id, _ := this.GetInt("id")
	server, err := service.ServerService.GetServer(id, service.SERVER_TYPE_NORMAL)
	this.checkError(err)

	if this.isPost() {
		valid := validation.Validation{}
		ip := this.GetString("server_ip")
		server.Area = this.GetString("area")
		server.Description = this.GetString("description")
		valid.Required(ip, "ip").Message("请输入服务器IP")
		valid.IP(ip, "ip").Message("服务器IP无效")
		if valid.HasErrors() {
			for _, err := range valid.Errors {
				this.showMsg(err.Message, MSG_ERR)
			}
		}
		server.Ip = ip
		err := service.ServerService.UpdateServer(server)
		this.checkError(err)
		this.redirect(beego.URLFor("ServerController.List"))
	}

	this.Data["pageTitle"] = "编辑服务器"
	this.Data["server"] = server
	this.display()
}

// 删除
func (this *ServerController) Del() {
	id, _ := this.GetInt("id")

	_, err := service.ServerService.GetServer(id, service.SERVER_TYPE_NORMAL)
	this.checkError(err)

	err = service.ServerService.DeleteServer(id)
	this.checkError(err)

	this.redirect(beego.URLFor("ServerController.List"))
}

// 项目列表
func (this *ServerController) Projects() {
	id, _ := this.GetInt("id")
	server, err := service.ServerService.GetServer(id, service.SERVER_TYPE_NORMAL)
	this.checkError(err)
	envList, err := service.EnvService.GetEnvListByServerId(id)
	this.checkError(err)

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

	this.Data["list"] = result
	this.Data["server"] = server
	this.Data["pageTitle"] = server.Ip + " 下的项目列表"
	this.display()
}
