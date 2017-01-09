package controllers

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/lisijie/gopub/app/entity"
	"github.com/lisijie/gopub/app/service"
	"strconv"
)

type EnvController struct {
	BaseController
}

func (c *EnvController) List() {
	projectId, _ := c.GetInt("project_id")
	envList, _ := service.EnvService.GetEnvListByProjectId(projectId)
	c.Data["pageTitle"] = "发布环境配置"
	c.Data["projectId"] = projectId
	c.Data["envList"] = envList
	c.display()
}

func (c *EnvController) Add() {
	projectId, _ := c.GetInt("project_id")

	project, err := service.ProjectService.GetProject(projectId)
	c.checkError(err)

	if c.isPost() {
		env := new(entity.Env)
		env.ProjectId = project.Id
		env.Name = c.GetString("name")
		env.SshUser = c.GetString("ssh_user")
		env.SshPort = c.GetString("ssh_port")
		env.SshKey = c.GetString("ssh_key")
		env.PubDir = c.GetString("pub_dir")
		env.BeforeShell = c.GetString("before_shell")
		env.AfterShell = c.GetString("after_shell")
		env.SendMail, _ = c.GetInt("send_mail")
		env.MailTplId, _ = c.GetInt("mail_tpl_id")
		env.MailTo = c.GetString("mail_to")
		env.MailCc = c.GetString("mail_cc")

		if env.Name == "" || env.SshUser == "" || env.SshPort == "" || env.SshKey == "" || env.PubDir == "" {
			c.showMsg("环境名称、SSH帐号、SSH端口、SSH KEY路径、发布目录不能为空。", MSG_ERR)
		}

		serverIds := c.GetStrings("serverIds")
		if len(serverIds) < 1 {
			c.showMsg("请选择服务器", MSG_ERR)
		}

		if env.SendMail > 0 {
			if env.MailTplId == 0 {
				c.showMsg("请选择邮件模板", MSG_ERR)
			}
		}

		env.ServerList = make([]entity.Server, 0, len(serverIds))
		for _, v := range serverIds {
			if sid, _ := strconv.Atoi(v); sid > 0 {
				if sv, err := service.ServerService.GetServer(sid); err == nil {
					env.ServerList = append(env.ServerList, *sv)
				} else {
					c.showMsg("服务器ID不存在: "+v, MSG_ERR)
				}
			}
		}
		if err := service.EnvService.AddEnv(env); err != nil {
			c.checkError(err)
		}

		c.redirect(beego.URLFor("EnvController.List", "project_id", projectId))
	}

	c.Data["serverList"], _ = service.ServerService.GetServerList(1, -1)
	c.Data["mailTplList"], _ = service.MailService.GetMailTplList()
	c.Data["project"] = project
	c.Data["pageTitle"] = "添加发布环境"
	c.display()
}

func (c *EnvController) Edit() {
	id, _ := c.GetInt("id")

	env, err := service.EnvService.GetEnv(id)
	c.checkError(err)

	if c.isPost() {
		env.Name = c.GetString("name")
		env.SshUser = c.GetString("ssh_user")
		env.SshPort = c.GetString("ssh_port")
		env.SshKey = c.GetString("ssh_key")
		env.PubDir = c.GetString("pub_dir")
		env.BeforeShell = c.GetString("before_shell")
		env.AfterShell = c.GetString("after_shell")
		env.SendMail, _ = c.GetInt("send_mail")
		env.MailTplId, _ = c.GetInt("mail_tpl_id")
		env.MailTo = c.GetString("mail_to")
		env.MailCc = c.GetString("mail_cc")

		if env.Name == "" || env.SshUser == "" || env.SshPort == "" || env.SshKey == "" || env.PubDir == "" {
			c.showMsg("环境名称、SSH帐号、SSH端口、SSH KEY路径、发布目录不能为空。", MSG_ERR)
		}

		serverIds := c.GetStrings("serverIds")
		if len(serverIds) < 1 {
			c.showMsg("请选择服务器", MSG_ERR)
		}

		if env.SendMail > 0 {
			if env.MailTplId == 0 {
				c.showMsg("请选择邮件模板", MSG_ERR)
			}
		}

		env.ServerList = make([]entity.Server, 0, len(serverIds))
		for _, v := range serverIds {
			if sid, _ := strconv.Atoi(v); sid > 0 {
				if sv, err := service.ServerService.GetServer(sid); err == nil {
					env.ServerList = append(env.ServerList, *sv)
				} else {
					c.showMsg("服务器ID不存在: "+v, MSG_ERR)
				}
			}
		}

		service.EnvService.SaveEnv(env)

		c.redirect(beego.URLFor("EnvController.List", "project_id", env.ProjectId))
	}

	serverList, _ := service.ServerService.GetServerList(1, -1)

	serverIds := make([]int, 0, len(env.ServerList))
	for _, v := range env.ServerList {
		serverIds = append(serverIds, v.Id)
	}

	jsonData, err := json.Marshal(serverIds)
	c.checkError(err)
	mailTplList, _ := service.MailService.GetMailTplList()

	c.Data["serverList"] = serverList
	c.Data["mailTplList"] = mailTplList
	c.Data["serverIds"] = string(jsonData)
	c.Data["env"] = env
	c.Data["pageTitle"] = "编辑发布环境"
	c.display()
}

func (c *EnvController) Del() {
	id, _ := c.GetInt("id")
	service.EnvService.DeleteEnv(id)
	c.redirect(beego.URLFor("EnvController.List", "project_id", id))
}
