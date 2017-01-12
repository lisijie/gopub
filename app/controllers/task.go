package controllers

import (
    "fmt"
    "github.com/astaxie/beego"
    "github.com/lisijie/gopub/app/entity"
    "github.com/lisijie/gopub/app/libs/utils"
    "github.com/lisijie/gopub/app/service"
)

type TaskController struct {
    BaseController
}

// 列表
func (c *TaskController) List() {
    status, _ := c.GetInt("status")
    page, _ := c.GetInt("page")
    startDate := c.GetString("start_date")
    endDate := c.GetString("end_date")
    projectId, _ := c.GetInt("project_id")
    if page < 1 {
        page = 1
    }
    filter := make([]interface{}, 0, 6)
    if projectId > 0 {
        filter = append(filter, "project_id", projectId)
    }
    if startDate != "" {
        filter = append(filter, "start_date", startDate)
    }
    if endDate != "" {
        filter = append(filter, "end_date", endDate)
    }
    if status == 1 {
        filter = append(filter, "pub_status", 3)
    } else {
        filter = append(filter, "pub_status__lt", 3)
    }

    list, count := service.TaskService.GetList(page, c.pageSize, filter...)
    projectList, _ := service.ProjectService.GetAllProject()

    c.Data["pageTitle"] = "发布单列表"
    c.Data["status"] = status
    c.Data["count"] = count
    c.Data["list"] = list
    c.Data["projectList"] = projectList
    c.Data["pageBar"] = utils.NewPager(page, int(count), c.pageSize, beego.URLFor("TaskController.List", "status", status, "project_id", projectId, "start_date", startDate, "end_date", endDate), true).ToString()
    c.Data["projectId"] = projectId
    c.Data["startDate"] = startDate
    c.Data["endDate"] = endDate
    c.display()
}

// 新建发布单
func (c *TaskController) Create() {

    if c.isPost() {
        projectId, _ := c.GetInt("project_id")
        envId, _ := c.GetInt("envId")
        verType, _ := c.GetInt("ver_type")
        startVer := c.GetString("start_ver")
        endVer := c.GetString("end_ver")
        message := c.GetString("editor_content")
        if envId < 1 {
            c.showMsg("请选择发布环境", MSG_ERR)
        }
        if verType == 2 {
            startVer = ""
        } else {
            repo, _ := service.ProjectService.GetRepository(projectId)
            if files, _ := repo.GetChangeFiles(startVer, endVer); len(files) < 1 {
                c.showMsg("版本区间 " + startVer + "..." + endVer + " 似乎没有差异文件！", MSG_ERR)
            }
        }

        project, err := service.ProjectService.GetProject(projectId)
        c.checkError(err)

        task := new(entity.Task)
        task.ProjectId = project.Id
        task.StartVer = startVer
        task.EndVer = endVer
        task.Message = message
        task.UserId = c.userId
        task.UserName = c.auth.GetUser().UserName
        task.PubEnvId = envId

        err = service.TaskService.AddTask(task)
        c.checkError(err)

        // 构建任务
        go service.TaskService.BuildTask(task)

        service.ActionService.Add("create_task", c.auth.GetUserName(), "task", task.Id, "")

        c.redirect(beego.URLFor("TaskController.List"))
    }

    projectId, _ := c.GetInt("project_id")
    c.Data["pageTitle"] = "新建发布单"

    if projectId < 1 {
        projectList, _ := service.ProjectService.GetAllProject()
        c.Data["list"] = projectList
        c.display("task/create_step1")
    } else {
        envList, _ := service.EnvService.GetEnvListByProjectId(projectId)
        c.Data["projectId"] = projectId
        c.Data["envList"] = envList
        c.display()
    }
}

// 列表
func (c *TaskController) GetRefs() {
    projectId, _ := c.GetInt("project_id")
    repo, err := service.ProjectService.GetRepository(projectId)
    c.checkError(err)
    err = repo.Update()
    c.checkError(err)
    tags, err := repo.GetTags()
    c.checkError(err)
    branches, err := repo.GetBranches()
    c.checkError(err)

    out := make(map[string]interface{})
    out["tags"] = tags
    out["branches"] = branches
    c.jsonResult(out)
}

// 任务详情
func (c *TaskController) Detail() {
    taskId, _ := c.GetInt("id")
    task, err := service.TaskService.GetTask(taskId)
    c.checkError(err)
    env, err := service.EnvService.GetEnv(task.PubEnvId)
    c.checkError(err)
    review, err := service.TaskService.GetReviewInfo(taskId)
    if err != nil {
        review = new(entity.TaskReview)
    }

    c.Data["env"] = env
    c.Data["task"] = task
    c.Data["review"] = review
    c.Data["pageTitle"] = "发布单详情"
    c.display()
}

// 获取状态
func (c *TaskController) GetStatus() {
    taskId, _ := c.GetInt("id")
    tp := c.GetString("type")

    task, err := service.TaskService.GetTask(taskId)
    c.checkError(err)

    out := make(map[string]interface{})
    switch tp {
    case "pub":
        out["status"] = task.PubStatus
        if task.PubStatus < 0 {
            out["msg"] = task.ErrorMsg
        } else {
            out["msg"] = task.PubLog
        }

    default:
        out["status"] = task.BuildStatus
        out["msg"] = task.ErrorMsg
    }

    c.jsonResult(out)
}

// 发布
func (c *TaskController) Publish() {
    taskId, _ := c.GetInt("id")
    step, _ := c.GetInt("step")
    if step < 1 {
        step = 1
    }
    task, err := service.TaskService.GetTask(taskId)
    c.checkError(err)

    if task.BuildStatus != 1 {
        c.showMsg("该任务单尚未构建成功！", MSG_ERR)
    }

    if task.ReviewStatus != 1 {
        c.showMsg("该任务单尚未通过审批！", MSG_ERR)
    }

    if task.PubStatus != 0 {
        step = 2
    }
    if task.PubStatus == 3 {
        step = 3
    }

    serverList, err := service.EnvService.GetEnvServers(task.PubEnvId)
    c.checkError(err)
    env, err := service.EnvService.GetEnv(task.PubEnvId)
    c.checkError(err)

    c.Data["serverList"] = serverList
    c.Data["task"] = task
    c.Data["env"] = env
    c.Data["pageTitle"] = "发布"

    c.display(fmt.Sprintf("task/publish-step%d", step))
}

// 开始发布
func (c *TaskController) StartPub() {
    taskId, _ := c.GetInt("id")

    if !c.auth.HasAccessPerm(c.controllerName, "publish") {
        c.showMsg("您没有执行该操作的权限", MSG_ERR)
    }

    err := service.DeployService.DeployTask(taskId)
    c.checkError(err)

    service.ActionService.Add("pub_task", c.auth.GetUserName(), "task", taskId, "")

    c.showMsg("", MSG_OK)
}

// 删除发布单
func (c *TaskController) Del() {
    taskId, _ := c.GetInt("id")
    refer := c.Ctx.Request.Referer()

    err := service.TaskService.DeleteTask(taskId)
    c.checkError(err)

    service.ActionService.Add("del_task", c.auth.GetUserName(), "task", taskId, "")

    if refer != "" {
        c.redirect(refer)
    } else {
        c.redirect(beego.URLFor("TaskController.List"))
    }
}
