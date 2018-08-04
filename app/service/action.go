package service

import (
	"fmt"
	"../entity"
)

// 系统动态
type actionService struct{}

func (this *actionService) table() string {
	return tableName("action")
}

// 添加记录
func (this *actionService) Add(action, actor, objectType string, objectId int, extra string) bool {
	act := new(entity.Action)
	act.Action = action
	act.Actor = actor
	act.ObjectType = objectType
	act.ObjectId = objectId
	act.Extra = extra
	o.Insert(act)
	return true
}

// 登录动态
func (this *actionService) Login(userName string, userId int, ip string) {
	this.Add("login", userName, "user", userId, ip)
}

// 退出登录
func (this *actionService) Logout(userName string, userId int, ip string) {
	this.Add("logout", userName, "user", userId, ip)
}

// 更新个人信息
func (this *actionService) UpdateProfile(userName string, userId int) {
	this.Add("update_profile", userName, "user", userId, "")
}

// 获取动态列表
func (this *actionService) GetList(page, pageSize int) ([]entity.Action, error) {
	var list []entity.Action
	num, err := o.QueryTable(this.table()).OrderBy("-create_time").Offset((page - 1) * pageSize).Limit(pageSize).All(&list)
	if num > 0 && err == nil {
		for i := 0; i < int(num); i++ {
			this.format(&list[i])
		}
	}
	return list, err
}

// 格式化
func (this *actionService) format(action *entity.Action) {
	switch action.Action {
	case "login":
		action.Message = fmt.Sprintf("<b>%s</b> 登录系统，IP为 <b>%s</b>。", action.Actor, action.Extra)
	case "logout":
		action.Message = fmt.Sprintf("<b>%s</b> 退出系统。", action.Actor)
	case "update_profile":
		action.Message = fmt.Sprintf("<b>%s</b> 更新了个人资料。", action.Actor)
	case "create_task":
		action.Message = fmt.Sprintf("<b>%s</b> 创建了编号为 <b class='blue'>%d</b> 的发布单。", action.Actor, action.ObjectId)
	case "pub_task":
		task, err := TaskService.GetTask(action.ObjectId)
		if err != nil {
			action.Message = fmt.Sprintf("<b>%s</b> 发布了编号为 <b class='blue'>%d</b> 版本。", action.Actor, action.ObjectId)
		} else {
			action.Message = fmt.Sprintf("<b>%s</b> 发布了 <span class='blue'>%s</span> 的 <b>%s</b> 版本。", action.Actor, task.ProjectInfo.Name, task.EndVer)
		}
	case "del_task":
		action.Message = fmt.Sprintf("<b>%s</b> 删除了编号为 <b class='blue'>%d</b> 的发布单。", action.Actor, action.ObjectId)
	case "review_task":
		task, err := TaskService.GetTask(action.ObjectId)
		if err != nil {
			if action.Extra == "1" {
				action.Message = fmt.Sprintf("<b>%s</b> 审批了编号为 <b class='blue'>%d</b> 的发布单，结果为：<b class='green'>通过</b>", action.Actor, action.ObjectId)
			} else {
				action.Message = fmt.Sprintf("<b>%s</b> 审批了编号为 <b class='blue'>%d</b> 的发布单，结果为：<b class='red'>不通过</b>", action.Actor, action.ObjectId)
			}
		} else {
			if action.Extra == "1" {
				action.Message = fmt.Sprintf("<b>%s</b> 审批了 <span class='text-primary'>%s</span> 编号为<b>%d</b>的发布单，结果为：<b class='green'>通过</b>", action.Actor, task.ProjectInfo.Name, action.ObjectId)
			} else if action.Extra == "-1" {
				action.Message = fmt.Sprintf("<b>%s</b> 审批了 <span class='text-primary'>%s</span> 编号为<b>%d</b>的发布单，结果为：<b class='red'>不通过</b>", action.Actor, task.ProjectInfo.Name, action.ObjectId)
			}
		}
	}
}
