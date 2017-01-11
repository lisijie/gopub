package controllers

import (
    "fmt"
    "github.com/astaxie/beego"
    "github.com/lisijie/gopub/app/entity"
    "github.com/lisijie/gopub/app/libs/utils"
    "github.com/lisijie/gopub/app/libs/ssh"
    "github.com/lisijie/gopub/app/service"
    "strconv"
    "strings"
)

type ProjectController struct {
    BaseController
}

// 项目列表
func (c *ProjectController) List() {
    page, _ := strconv.Atoi(c.GetString("page"))
    if page < 1 {
        page = 1
    }

    count, _ := service.ProjectService.GetTotal()
    list, _ := service.ProjectService.GetList(page, c.pageSize)

    c.Data["count"] = count
    c.Data["list"] = list
    c.Data["pageBar"] = utils.NewPager(page, int(count), c.pageSize, beego.URLFor("ProjectController.List"), true).ToString()
    c.Data["pageTitle"] = "项目列表"
    c.display()
}

// 添加项目
func (c *ProjectController) Add() {

    if c.isPost() {
        p := &entity.Project{}
        p.Name = c.GetString("project_name")
        p.Domain = c.GetString("project_domain")
        p.RepoUrl = c.GetString("repo_url")
        p.AgentId, _ = c.GetInt("agent_id")
        p.IgnoreList = c.GetString("ignore_list")
        p.BeforeShell = c.GetString("before_shell")
        p.AfterShell = c.GetString("after_shell")
        p.TaskReview, _ = c.GetInt("task_review")
        if v, _ := c.GetInt("create_verfile"); v > 0 {
            p.CreateVerfile = 1
        } else {
            p.CreateVerfile = 0
        }
        p.VerfilePath = strings.Replace(c.GetString("verfile_path"), ".", "", -1)

        if err := c.validProject(p); err != nil {
            c.showMsg(err.Error(), MSG_ERR)
        }

        err := service.ProjectService.AddProject(p)
        c.checkError(err)

        // 克隆仓库
        go service.ProjectService.CloneRepo(p.Id)

        service.ActionService.Add("add_project", c.auth.GetUserName(), "project", p.Id, "")

        c.redirect(beego.URLFor("ProjectController.List"))
    }

    agentList, err := service.ServerService.GetAgentList(1, -1)
    c.checkError(err)
    c.Data["pageTitle"] = "添加项目"
    c.Data["agentList"] = agentList
    c.display()
}

// 编辑项目
func (c *ProjectController) Edit() {
    id, _ := c.GetInt("id")
    p, err := service.ProjectService.GetProject(id)
    c.checkError(err)

    if c.isPost() {
        p.Name = c.GetString("project_name")
        p.AgentId, _ = c.GetInt("agent_id")
        p.IgnoreList = c.GetString("ignore_list")
        p.BeforeShell = c.GetString("before_shell")
        p.AfterShell = c.GetString("after_shell")
        p.TaskReview, _ = c.GetInt("task_review")
        if p.Status == -1 {
            p.RepoUrl = c.GetString("repo_url")
        }
        if v, _ := c.GetInt("create_verfile"); v > 0 {
            p.CreateVerfile = 1
        } else {
            p.CreateVerfile = 0
        }
        p.VerfilePath = strings.Replace(c.GetString("verfile_path"), ".", "", -1)

        if err := c.validProject(p); err != nil {
            c.showMsg(err.Error(), MSG_ERR)
        }

        err := service.ProjectService.UpdateProject(p, "Name", "AgentId", "IgnoreList", "BeforeShell", "AfterShell", "RepoUrl", "CreateVerfile", "VerfilePath", "TaskReview")
        c.checkError(err)

        service.ActionService.Add("edit_project", c.auth.GetUserName(), "project", p.Id, "")

        c.redirect(beego.URLFor("ProjectController.List"))
    }

    agentList, err := service.ServerService.GetAgentList(1, -1)
    c.checkError(err)

    c.Data["project"] = p
    c.Data["agentList"] = agentList
    c.Data["pageTitle"] = "编辑项目"
    c.display()
}

// 删除项目
func (c *ProjectController) Del() {
    id, _ := c.GetInt("id")

    err := service.ProjectService.DeleteProject(id)
    c.checkError(err)

    service.ActionService.Add("del_project", c.auth.GetUserName(), "project", id, "")

    c.redirect(beego.URLFor("ProjectController.List"))
}

// 重新克隆
func (c *ProjectController) Clone() {
    id, _ := c.GetInt("id")
    project, err := service.ProjectService.GetProject(id)
    c.checkError(err)
    if project.Status != -1 {
        c.showMsg("只能对克隆失败的项目操作.", MSG_ERR)
    }

    project.Status = 0
    service.ProjectService.UpdateProject(project, "Status")
    go service.ProjectService.CloneRepo(id)

    c.showMsg("", MSG_OK)
}

// 获取项目克隆状态
func (c *ProjectController) GetStatus() {
    id, _ := c.GetInt("id")
    project, _ := service.ProjectService.GetProject(id)

    out := make(map[string]interface{})
    out["status"] = project.Status
    out["error"] = project.ErrorMsg

    c.jsonResult(out)
}

// 验证提交
func (c *ProjectController) validProject(p *entity.Project) error {
    errorMsg := ""
    if p.Name == "" {
        errorMsg = "请输入项目名称"
    } else if p.Domain == "" {
        errorMsg = "请输入项目标识"
    } else if p.RepoUrl == "" {
        errorMsg = "请输入仓库地址"
    } else if p.AgentId == 0 {
        errorMsg = "请选择跳板机"
    } else {
        agent, err := service.ServerService.GetServer(p.AgentId)
        if err != nil {
            return err
        }
        serv := ssh.NewServerConn(&ssh.Config{
            Addr:fmt.Sprintf("%s:%d", agent.Ip, agent.SshPort),
            User:agent.SshUser,
            Password:agent.SshPwd,
            Key:agent.SshKey,
        })
        workPath := fmt.Sprintf("%s/%s", agent.WorkDir, p.Domain)

        if err := serv.TryConnect(); err != nil {
            errorMsg = "无法连接到跳板机: " + err.Error()
        } else if _, err := serv.RunCmd("mkdir -p " + workPath); err != nil {
            errorMsg = "无法创建跳板机工作目录: " + err.Error()
        }
        serv.Close()
    }

    if errorMsg != "" {
        return fmt.Errorf(errorMsg)
    }
    return nil
}
