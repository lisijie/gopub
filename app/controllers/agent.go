package controllers

import (
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/validation"
	"github.com/lisijie/gopub/app/entity"
	"github.com/lisijie/gopub/app/libs/utils"
	"github.com/lisijie/gopub/app/libs/ssh"
	"github.com/lisijie/gopub/app/service"
	"strconv"
)

type AgentController struct {
	BaseController
}

// 列表
func (c *AgentController) List() {
	page, _ := strconv.Atoi(c.GetString("page"))
	if page < 1 {
		page = 1
	}
	count, err := service.ServerService.GetTotal(service.SERVER_TYPE_AGENT)
	c.checkError(err)
	serverList, err := service.ServerService.GetAgentList(page, c.pageSize)
	c.checkError(err)

	c.Data["count"] = count
	c.Data["list"] = serverList
	c.Data["pageBar"] = utils.NewPager(page, int(count), c.pageSize, beego.URLFor("AgentController.List"), true).ToString()
	c.Data["pageTitle"] = "跳板机列表"
	c.display()
}

// 添加
func (c *AgentController) Add() {
	if c.isPost() {
		server := &entity.Server{}
		server.TypeId = service.SERVER_TYPE_AGENT
		server.Ip = c.GetString("server_ip")
		server.Area = c.GetString("area")
		server.SshPort, _ = c.GetInt("ssh_port")
		server.SshUser = c.GetString("ssh_user")
		server.SshPwd = c.GetString("ssh_pwd")
		server.SshKey = c.GetString("ssh_key")
		server.WorkDir = c.GetString("work_dir")
		server.Description = c.GetString("description")
		err := c.validServer(server)
		c.checkError(err)
		err = service.ServerService.AddServer(server)
		c.checkError(err)
		//service.ActionService.Add("add_agent", c.auth.GetUserName(), "server", server.Id, server.Ip)
		c.redirect(beego.URLFor("AgentController.List"))
	}

	c.Data["pageTitle"] = "添加跳板机"
	c.display()
}

// 编辑
func (c *AgentController) Edit() {
	id, _ := c.GetInt("id")
	server, err := service.ServerService.GetServer(id, service.SERVER_TYPE_AGENT)
	c.checkError(err)

	if c.isPost() {
		server.Ip = c.GetString("server_ip")
		server.Area = c.GetString("area")
		server.SshPort, _ = c.GetInt("ssh_port")
		server.SshUser = c.GetString("ssh_user")
		server.SshPwd = c.GetString("ssh_pwd")
		server.SshKey = c.GetString("ssh_key")
		server.WorkDir = c.GetString("work_dir")
		server.Description = c.GetString("description")
		err := c.validServer(server)
		c.checkError(err)
		err = service.ServerService.UpdateServer(server)
		c.checkError(err)
		//service.ActionService.Add("edit_agent", c.auth.GetUserName(), "server", server.Id, server.Ip)
		c.redirect(beego.URLFor("AgentController.List"))
	}

	c.Data["pageTitle"] = "编辑跳板机"
	c.Data["server"] = server
	c.display()
}

// 删除
func (c *AgentController) Del() {
	id, _ := c.GetInt("id")

	_, err := service.ServerService.GetServer(id, service.SERVER_TYPE_AGENT)
	c.checkError(err)

	err = service.ServerService.DeleteServer(id)
	c.checkError(err)
	//service.ActionService.Add("del_agent", c.auth.GetUserName(), "server", id, "")

	c.redirect(beego.URLFor("AgentController.List"))
}

// 项目列表
func (c *AgentController) Projects() {
	id, _ := c.GetInt("id")
	server, err := service.ServerService.GetServer(id, service.SERVER_TYPE_AGENT)
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

func (c *AgentController) validServer(server *entity.Server) error {
	valid := validation.Validation{}
	valid.Required(server.Ip, "ip").Message("请输入服务器IP")
	valid.Range(server.SshPort, 1, 65535, "ssh_port").Message("SSH端口无效")
	valid.Required(server.SshUser, "ssh_user").Message("SSH用户名不能为空")
	valid.Required(server.WorkDir, "work_dir").Message("工作目录不能为空")
	valid.IP(server.Ip, "ip").Message("服务器IP无效")
	if valid.HasErrors() {
		for _, err := range valid.Errors {
			return errors.New(err.Message)
		}
	}
	if server.SshKey != "" && !utils.IsFile(utils.RealPath(server.SshKey)) {
		return errors.New("SSH Key不存在:" + server.SshKey)
	}

	addr := fmt.Sprintf("%s:%d", server.Ip, server.SshPort)
	serv := ssh.NewServerConn(addr, server.SshUser, server.SshKey)

	if err := serv.TryConnect(); err != nil {
		return errors.New("无法连接到跳板机: " + err.Error())
	} else if _, err := serv.RunCmd("mkdir -p " + server.WorkDir); err != nil {
		return errors.New("无法创建跳板机工作目录: " + err.Error())
	}
	serv.Close()

	return nil
}
