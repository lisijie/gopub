package controllers

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/lisijie/gopub/app/service"
	"os"
	"runtime"
	"time"
)

type MainController struct {
	BaseController
}

// 首页
func (c *MainController) Index() {
	c.Data["pageTitle"] = "系统概况"

	projectsStat := service.TaskService.GetProjectPubStat()
	popProjects := make([]map[string]interface{}, 0, 4)
	for k, stat := range projectsStat {
		projectInfo, err := service.ProjectService.GetProject(stat["project_id"])
		if err != nil {
			continue
		}
		if k > 4 {
			break
		}
		info := make(map[string]interface{})
		info["project_name"] = projectInfo.Name
		info["version"] = projectInfo.Version
		info["version_time"] = beego.Date(projectInfo.VersionTime, "Y-m-d H:i:s")
		info["count"] = stat["count"]
		popProjects = append(popProjects, info)
	}

	feeds, _ := service.ActionService.GetList(1, 7)
	c.Data["feeds"] = feeds
	c.Data["serverNum"], _ = service.ServerService.GetTotal(service.SERVER_TYPE_NORMAL)
	c.Data["userNum"], _ = service.UserService.GetTotal()
	c.Data["projectNum"], _ = service.ProjectService.GetTotal()
	c.Data["pubNum"], _ = service.TaskService.GetPubTotal()
	c.Data["popProjects"] = popProjects
	c.Data["hostname"], _ = os.Hostname()
	c.Data["gover"] = runtime.Version()
	c.Data["os"] = runtime.GOOS
	c.Data["goroutineNum"] = runtime.NumGoroutine()
	c.Data["cpuNum"] = runtime.NumCPU()
	c.Data["arch"] = runtime.GOARCH
	c.Data["dbVerson"] = service.DBVersion()
	c.Data["dataDir"] = beego.AppConfig.String("data_dir")
	up, day, hour, min, sec := c.getUptime()
	c.Data["uptime"] = fmt.Sprintf("%s，已运行 %d天 %d小时 %d分钟 %d秒", beego.Date(up, "Y-m-d H:i:s"), day, hour, min, sec)
	c.display()
}

func (c *MainController) getUptime() (up time.Time, day, hour, min, sec int) {
	ts, _ := beego.AppConfig.Int64("up_time")
	up = time.Unix(ts, 0)
	uptime := int(time.Now().Sub(up) / time.Second)
	if uptime >= 86400 {
		day = uptime / 86400
		uptime %= 86400
	}
	if uptime >= 3600 {
		hour = uptime / 3600
		uptime %= 3600
	}
	if uptime >= 60 {
		min = uptime / 60
		uptime %= 60
	}
	sec = uptime
	return
}

// 发版统计
func (c *MainController) GetPubStat() {
	rangeType := c.GetString("range")
	result := service.TaskService.GetPubStat(rangeType)

	ticks := make([]interface{}, 0)
	chart := make([]interface{}, 0)
	json := make(map[string]interface{}, 0)
	switch rangeType {
	case "this_month":
		year, month, _ := time.Now().Date()
		maxDay := time.Date(year, month, 1, 0, 0, 0, 0, time.Local).AddDate(0, 1, 0).AddDate(0, 0, -1).Day()

		for i := 1; i <= maxDay; i++ {
			var row [3]interface{}
			row[0] = i
			row[1] = fmt.Sprintf("%02d", i)
			row[2] = fmt.Sprintf("%d-%02d-%02d", year, month, i)
			ticks = append(ticks, row)
			if v, ok := result[i]; ok {
				chart = append(chart, []int{i, v})
			} else {
				chart = append(chart, []int{i, 0})
			}
		}
	case "last_month":
		year, month, _ := time.Now().AddDate(0, -1, 0).Date()
		maxDay := time.Date(year, month, 1, 0, 0, 0, 0, time.Local).AddDate(0, 1, 0).AddDate(0, 0, -1).Day()

		for i := 1; i <= maxDay; i++ {
			var row [3]interface{}
			row[0] = i
			row[1] = fmt.Sprintf("%02d", i)
			row[2] = fmt.Sprintf("%d-%02d-%02d", year, month, i)
			ticks = append(ticks, row)
			if v, ok := result[i]; ok {
				chart = append(chart, []int{i, v})
			} else {
				chart = append(chart, []int{i, 0})
			}
		}
	case "this_year":
		year := time.Now().Year()
		for i := 1; i <= 12; i++ {
			var row [3]interface{}
			row[0] = i
			row[1] = fmt.Sprintf("%d月", i)
			row[2] = fmt.Sprintf("%d年%d月", year, i)
			ticks = append(ticks, row)
			if v, ok := result[i]; ok {
				chart = append(chart, []int{i, v})
			} else {
				chart = append(chart, []int{i, 0})
			}
		}
	case "last_year":
		year := time.Now().Year() - 1
		for i := 1; i <= 12; i++ {
			var row [3]interface{}
			row[0] = i
			row[1] = fmt.Sprintf("%d月", i)
			row[2] = fmt.Sprintf("%d年%d月", year, i)
			ticks = append(ticks, row)
			if v, ok := result[i]; ok {
				chart = append(chart, []int{i, v})
			} else {
				chart = append(chart, []int{i, 0})
			}
		}
	}

	json["ticks"] = ticks
	json["chart"] = chart
	c.Data["json"] = json
	c.ServeJSON()
}

// 个人信息
func (c *MainController) Profile() {
	beego.ReadFromRequest(&c.Controller)
	user := c.auth.GetUser()

	if c.isPost() {
		flash := beego.NewFlash()
		email := c.GetString("email")
		sex, _ := c.GetInt("sex")
		password1 := c.GetString("password1")
		password2 := c.GetString("password2")

		user.Email = email
		user.Sex = sex
		service.UserService.UpdateUser(user, "Email", "Sex")
		if password1 != "" {
			if len(password1) < 6 {
				flash.Error("密码长度必须大于6位")
				flash.Store(&c.Controller)
				c.redirect(beego.URLFor(".Profile"))
			} else if password2 != password1 {
				flash.Error("两次输入的密码不一致")
				flash.Store(&c.Controller)
				c.redirect(beego.URLFor(".Profile"))
			} else {
				service.UserService.ModifyPassword(c.userId, password1)
			}
		}
		service.ActionService.UpdateProfile(c.auth.GetUser().UserName, c.userId)
		flash.Success("修改成功！")
		flash.Store(&c.Controller)
		c.redirect(beego.URLFor(".Profile"))
	}

	c.Data["pageTitle"] = "个人信息"
	c.Data["user"] = user
	c.display()
}

// 登录
func (c *MainController) Login() {
	if c.userId > 0 {
		c.redirect("/")
	}
	beego.ReadFromRequest(&c.Controller)
	if c.isPost() {
		flash := beego.NewFlash()
		username := c.GetString("username")
		password := c.GetString("password")
		remember := c.GetString("remember")
		if username != "" && password != "" {
			token, err := c.auth.Login(username, password)
			if err != nil {
				flash.Error(err.Error())
				flash.Store(&c.Controller)
				c.redirect("/login")
			} else {
				if remember == "yes" {
					c.Ctx.SetCookie("auth", token, 7*86400)
				} else {
					c.Ctx.SetCookie("auth", token)
				}
				service.ActionService.Login(username, c.auth.GetUserId(), c.getClientIp())
				c.redirect(beego.URLFor(".Index"))
			}

		}
	}

	c.TplName = "main/login.html"
}

// 退出登录
func (c *MainController) Logout() {
	service.ActionService.Logout(c.auth.GetUser().UserName, c.auth.GetUserId(), c.getClientIp())
	c.auth.Logout()
	c.Ctx.SetCookie("auth", "")
	c.redirect(beego.URLFor(".Login"))
}
