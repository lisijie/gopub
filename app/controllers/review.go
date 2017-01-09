package controllers

import (
	"github.com/astaxie/beego"
	"github.com/lisijie/gopub/app/entity"
	"github.com/lisijie/gopub/app/libs/utils"
	"github.com/lisijie/gopub/app/service"
)

type ReviewController struct {
	BaseController
}

// 列表
func (c *ReviewController) List() {
	status, _ := c.GetInt("status")
	page, _ := c.GetInt("page")
	startDate := c.GetString("start_date")
	endDate := c.GetString("end_date")

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

	list, count := service.TaskService.GetList(page, c.pageSize, filter...)
	envList := make(map[int]*entity.Env)
	for k, v := range list {
		if _, ok := envList[v.PubEnvId]; !ok {
			envList[v.PubEnvId], _ = service.EnvService.GetEnv(v.PubEnvId)
		}
		list[k].EnvInfo = envList[v.PubEnvId]
	}

	c.Data["pageTitle"] = "审批列表"
	c.Data["status"] = status
	c.Data["count"] = count
	c.Data["list"] = list
	c.Data["pageBar"] = utils.NewPager(page, int(count), c.pageSize, beego.URLFor("ReviewController.List", "status", status, "start_date", startDate, "end_date", endDate), true).ToString()
	c.Data["startDate"] = startDate
	c.Data["endDate"] = endDate
	c.display()
}

// 审批
func (c *ReviewController) Review() {
	id, _ := c.GetInt("id")

	if c.isPost() {
		status, _ := c.GetInt("status")
		message := c.GetString("message")
		err := service.TaskService.ReviewTask(id, c.userId, status, message)
		c.checkError(err)
		service.ActionService.Add("review_task", c.auth.GetUserName(), "task", id, c.GetString("status"))
		c.redirect(beego.URLFor("ReviewController.List"))
	}

	task, err := service.TaskService.GetTask(id)
	c.checkError(err)
	env, err := service.EnvService.GetEnv(task.PubEnvId)
	c.checkError(err)

	c.Data["pageTitle"] = "审批发布单"
	c.Data["env"] = env
	c.Data["task"] = task
	c.display()
}

// 详情
func (c *ReviewController) Detail() {
	id, _ := c.GetInt("id")

	task, err := service.TaskService.GetTask(id)
	c.checkError(err)
	env, err := service.EnvService.GetEnv(task.PubEnvId)
	c.checkError(err)
	review, err := service.TaskService.GetReviewInfo(id)
	if err != nil {
		c.showMsg("审批记录不存在。", MSG_ERR)
	}

	c.Data["pageTitle"] = "浏览详情"
	c.Data["env"] = env
	c.Data["task"] = task
	c.Data["review"] = review
	c.Data["refer"] = c.Ctx.Request.Referer()
	c.display()
}
