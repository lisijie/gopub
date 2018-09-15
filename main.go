package main

import (
	"./app/controllers"
	_ "./app/mail"
	"./app/service"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/beego/i18n"
	"time"
)

const VERSION = "2.0.1"

func main() {
	service.Init()

	beego.AppConfig.Set("version", VERSION)
	if beego.AppConfig.String("runmode") == "dev" {
		beego.SetLevel(beego.LevelDebug)
	} else {
		beego.SetLevel(beego.LevelInformational)
		beego.SetLogger("file", `{"filename":"`+beego.AppConfig.String("log_file")+`"}`)
		beego.BeeLogger.DelLogger("console")
	}

	beego.Router("/", &controllers.MainController{}, "*:Index")
	beego.Router("/login", &controllers.MainController{}, "*:Login")
	beego.Router("/logout", &controllers.MainController{}, "*:Logout")
	beego.Router("/profile", &controllers.MainController{}, "*:Profile")

	beego.AutoRouter(&controllers.ProjectController{})
	beego.AutoRouter(&controllers.TaskController{})
	beego.AutoRouter(&controllers.ServerController{})
	beego.AutoRouter(&controllers.EnvController{})
	beego.AutoRouter(&controllers.UserController{})
	beego.AutoRouter(&controllers.RoleController{})
	beego.AutoRouter(&controllers.MailTplController{})
	beego.AutoRouter(&controllers.AgentController{})
	beego.AutoRouter(&controllers.ReviewController{})
	beego.AutoRouter(&controllers.MainController{})

	// 记录启动时间
	beego.AppConfig.Set("up_time", fmt.Sprintf("%d", time.Now().Unix()))

	beego.AddFuncMap("i18n", i18n.Tr)

	beego.SetStaticPath("/assets", "assets")
	beego.Run()
}
