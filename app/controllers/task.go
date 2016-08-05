package controllers

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/lisijie/gopub/app/entity"
	"github.com/lisijie/gopub/app/libs"
	"github.com/lisijie/gopub/app/service"
)

type TaskController struct {
	BaseController
}

// 列表
func (this *TaskController) List() {
	status, _ := this.GetInt("status")
	page, _ := this.GetInt("page")
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	projectId, _ := this.GetInt("project_id")
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

	list, count := service.TaskService.GetList(page, this.pageSize, filter...)
	projectList, _ := service.ProjectService.GetAllProject()

	this.Data["pageTitle"] = "发布单列表"
	this.Data["status"] = status
	this.Data["count"] = count
	this.Data["list"] = list
	this.Data["projectList"] = projectList
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("TaskController.List", "status", status, "project_id", projectId, "start_date", startDate, "end_date", endDate), true).ToString()
	this.Data["projectId"] = projectId
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.display()
}

// 新建发布单
func (this *TaskController) Create() {

	if this.isPost() {
		projectId, _ := this.GetInt("project_id")
		envId, _ := this.GetInt("envId")
		verType, _ := this.GetInt("ver_type")
		startVer := this.GetString("start_ver")
		endVer := this.GetString("end_ver")
		message := this.GetString("editor_content")
		if envId < 1 {
			this.showMsg("请选择发布环境", MSG_ERR)
		}
		if verType == 2 {
			startVer = ""
		} else {
			if libs.VerCompare(startVer, endVer) != -1 {
				this.showMsg("起始版本必须小于结束版本", MSG_ERR)
			} else {
				repo, _ := service.RepositoryService.GetRepoByProjectId(projectId)
				if count, _ := repo.GetDiffFileCount(startVer, endVer); count < 1 {
					this.showMsg("版本区间 "+startVer+" - "+endVer+" 似乎没有差异文件！", MSG_ERR)
				}
			}
		}

		project, err := service.ProjectService.GetProject(projectId)
		this.checkError(err)

		task := new(entity.Task)
		task.ProjectId = project.Id
		task.StartVer = startVer
		task.EndVer = endVer
		task.Message = message
		task.UserId = this.userId
		task.UserName = this.auth.GetUser().UserName
		task.PubEnvId = envId

		err = service.TaskService.AddTask(task)
		this.checkError(err)

		// 构建任务
		go service.TaskService.BuildTask(task)

		service.ActionService.Add("create_task", this.auth.GetUserName(), "task", task.Id, "")

		this.redirect(beego.URLFor("TaskController.List"))
	}

	projectId, _ := this.GetInt("project_id")
	this.Data["pageTitle"] = "新建发布单"

	if projectId < 1 {
		projectList, _ := service.ProjectService.GetAllProject()
		this.Data["list"] = projectList
		this.display("task/create_step1")
	} else {
		envList, _ := service.EnvService.GetEnvListByProjectId(projectId)
		this.Data["projectId"] = projectId
		this.Data["envList"] = envList
		this.display()
	}
}

// 标签列表
func (this *TaskController) GetTags() {
	projectId, _ := this.GetInt("project_id")

	list, err := service.RepositoryService.GetTags(projectId, 10)
	this.checkError(err)

	out := make(map[string]interface{})
	out["list"] = list
	this.jsonResult(out)
}

// 任务详情
func (this *TaskController) Detail() {
	taskId, _ := this.GetInt("id")
	task, err := service.TaskService.GetTask(taskId)
	this.checkError(err)
	env, err := service.EnvService.GetEnv(task.PubEnvId)
	this.checkError(err)
	review, err := service.TaskService.GetReviewInfo(taskId)
	if err != nil {
		review = new(entity.TaskReview)
	}

	this.Data["env"] = env
	this.Data["task"] = task
	this.Data["review"] = review
	this.Data["pageTitle"] = "发布单详情"
	this.display()
}

// 获取状态
func (this *TaskController) GetStatus() {
	taskId, _ := this.GetInt("id")
	tp := this.GetString("type")

	task, err := service.TaskService.GetTask(taskId)
	this.checkError(err)

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

	this.jsonResult(out)
}

// 发布
func (this *TaskController) Publish() {
	taskId, _ := this.GetInt("id")
	step, _ := this.GetInt("step")
	if step < 1 {
		step = 1
	}
	task, err := service.TaskService.GetTask(taskId)
	this.checkError(err)

	if task.BuildStatus != 1 {
		this.showMsg("该任务单尚未构建成功！", MSG_ERR)
	}

	if task.ReviewStatus != 1 {
		this.showMsg("该任务单尚未通过审批！", MSG_ERR)
	}

	if task.PubStatus != 0 {
		step = 2
	}
	if task.PubStatus == 3 {
		step = 3
	}

	serverList, err := service.EnvService.GetEnvServers(task.PubEnvId)
	this.checkError(err)
	env, err := service.EnvService.GetEnv(task.PubEnvId)
	this.checkError(err)

	this.Data["serverList"] = serverList
	this.Data["task"] = task
	this.Data["env"] = env
	this.Data["pageTitle"] = "发布"

	this.display(fmt.Sprintf("task/publish-step%d", step))
}

// 开始发布
func (this *TaskController) StartPub() {
	taskId, _ := this.GetInt("id")

	if !this.auth.HasAccessPerm(this.controllerName, "publish") {
		this.showMsg("您没有执行该操作的权限", MSG_ERR)
	}

	err := service.DeployService.DeployTask(taskId)
	this.checkError(err)

	service.ActionService.Add("pub_task", this.auth.GetUserName(), "task", taskId, "")

	this.showMsg("", MSG_OK)
}

// 删除发布单
func (this *TaskController) Del() {
	taskId, _ := this.GetInt("id")
	refer := this.Ctx.Request.Referer()

	err := service.TaskService.DeleteTask(taskId)
	this.checkError(err)

	service.ActionService.Add("del_task", this.auth.GetUserName(), "task", taskId, "")

	if refer != "" {
		this.redirect(refer)
	} else {
		this.redirect(beego.URLFor("TaskController.List"))
	}
}
