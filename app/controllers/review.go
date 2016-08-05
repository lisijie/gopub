package controllers

import (
	"github.com/astaxie/beego"
	"github.com/lisijie/gopub/app/entity"
	"github.com/lisijie/gopub/app/libs"
	"github.com/lisijie/gopub/app/service"
)

type ReviewController struct {
	BaseController
}

// 列表
func (this *ReviewController) List() {
	status, _ := this.GetInt("status")
	page, _ := this.GetInt("page")
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")

	if page < 1 {
		page = 1
	}
	filter := make([]interface{}, 0, 6)
	if status == 0 {
		filter = append(filter, "review_status", 0)
	} else {
		filter = append(filter, "review_status__in", []int{1, -1})
	}
	if startDate != "" {
		filter = append(filter, "start_date", startDate)
	}
	if endDate != "" {
		filter = append(filter, "end_date", endDate)
	}

	list, count := service.TaskService.GetList(page, this.pageSize, filter...)
	envList := make(map[int]*entity.Env)
	for k, v := range list {
		if _, ok := envList[v.PubEnvId]; !ok {
			envList[v.PubEnvId], _ = service.EnvService.GetEnv(v.PubEnvId)
		}
		list[k].EnvInfo = envList[v.PubEnvId]
	}

	this.Data["pageTitle"] = "审批列表"
	this.Data["status"] = status
	this.Data["count"] = count
	this.Data["list"] = list
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("ReviewController.List", "status", status, "start_date", startDate, "end_date", endDate), true).ToString()
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.display()
}

// 审批
func (this *ReviewController) Review() {
	id, _ := this.GetInt("id")

	if this.isPost() {
		status, _ := this.GetInt("status")
		message := this.GetString("message")
		err := service.TaskService.ReviewTask(id, this.userId, status, message)
		this.checkError(err)
		service.ActionService.Add("review_task", this.auth.GetUserName(), "task", id, this.GetString("status"))
		this.redirect(beego.URLFor("ReviewController.List"))
	}

	task, err := service.TaskService.GetTask(id)
	this.checkError(err)
	env, err := service.EnvService.GetEnv(task.PubEnvId)
	this.checkError(err)

	this.Data["pageTitle"] = "审批发布单"
	this.Data["env"] = env
	this.Data["task"] = task
	this.display()
}

// 详情
func (this *ReviewController) Detail() {
	id, _ := this.GetInt("id")

	task, err := service.TaskService.GetTask(id)
	this.checkError(err)
	env, err := service.EnvService.GetEnv(task.PubEnvId)
	this.checkError(err)
	review, err := service.TaskService.GetReviewInfo(id)
	if err != nil {
		this.showMsg("审批记录不存在。", MSG_ERR)
	}

	this.Data["pageTitle"] = "浏览详情"
	this.Data["env"] = env
	this.Data["task"] = task
	this.Data["review"] = review
	this.Data["refer"] = this.Ctx.Request.Referer()
	this.display()
}
