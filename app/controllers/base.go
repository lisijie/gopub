package controllers

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/beego/i18n"
	"github.com/lisijie/gopub/app/service"
	"io/ioutil"
	"strings"
)

const (
	MSG_OK       = 0  // ajax输出错误码，成功
	MSG_ERR      = -1 // 错误
	MSG_REDIRECT = -2 // 重定向
)

type BaseController struct {
	beego.Controller
	auth           *service.AuthService // 验证服务
	userId         int                  // 当前登录的用户id
	controllerName string               // 控制器名
	actionName     string               // 动作名
	pageSize       int                  // 默认分页大小
	lang           string               // 当前语言环境
}

// 重写GetString方法，移除前后空格
func (c *BaseController) GetString(name string, def ...string) string {
	return strings.TrimSpace(c.Controller.GetString(name, def...))
}

func (c *BaseController) Prepare() {
	c.Ctx.Output.Header("X-Powered-By", "GoPub/"+beego.AppConfig.String("version"))
	c.Ctx.Output.Header("X-Author-By", "lisijie")
	controllerName, actionName := c.GetControllerAndAction()
	c.controllerName = strings.ToLower(controllerName[0 : len(controllerName)-10])
	c.actionName = strings.ToLower(actionName)
	c.pageSize = 20
	c.initAuth()
	c.initLang()
	c.getMenuList()
}

func (c *BaseController) initLang() {
	c.lang = "zh-CN"
	c.Data["lang"] = c.lang
	if !i18n.IsExist(c.lang) {
		if err := i18n.SetMessage(c.lang, beego.AppPath+"/conf/locale_"+c.lang+".ini"); err != nil {
			beego.Error("Fail to set message file: " + err.Error())
			return
		}

	}
}

//登录验证
func (c *BaseController) initAuth() {
	token := c.Ctx.GetCookie("auth")
	c.auth = service.NewAuth(token)
	c.userId = c.auth.GetUserId()

	if !c.auth.IsLogined() {
		if c.controllerName != "main" ||
			(c.controllerName == "main" && c.actionName != "logout" && c.actionName != "login") {
			c.redirect(beego.URLFor("MainController.Login"))
		}
	} else {
		if !c.auth.HasAccessPerm(c.controllerName, c.actionName) {
			c.showMsg("您没有执行该操作的权限", MSG_ERR)
		}
	}
}

//渲染模版
func (c *BaseController) display(tpl ...string) {
	var tplname string
	if len(tpl) > 0 {
		tplname = tpl[0] + ".html"
	} else {
		tplname = c.controllerName + "/" + c.actionName + ".html"
	}

	c.Layout = "layout/layout.html"
	c.TplName = tplname

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Header"] = "layout/sections/header.html"
	c.LayoutSections["Footer"] = "layout/sections/footer.html"
	c.LayoutSections["Navbar"] = "layout/sections/navbar.html"
	c.LayoutSections["Sidebar"] = "layout/sections/sidebar.html"

	user := c.auth.GetUser()

	c.Data["version"] = beego.AppConfig.String("version")
	c.Data["curRoute"] = c.controllerName + "." + c.actionName
	c.Data["loginUserId"] = user.Id
	c.Data["loginUserName"] = user.UserName
	c.Data["loginUserSex"] = user.Sex
	c.Data["menuList"] = c.getMenuList()
}

// 重定向
func (c *BaseController) redirect(url string) {
	if c.IsAjax() {
		c.showMsg("", MSG_REDIRECT, url)
	} else {
		c.Redirect(url, 302)
		c.StopRun()
	}
}

// 是否POST提交
func (c *BaseController) isPost() bool {
	return c.Ctx.Request.Method == "POST"
}

// 提示消息
func (c *BaseController) showMsg(msg string, code int, redirect ...string) {
	out := make(map[string]interface{})
	out["status"] = code
	out["msg"] = msg
	out["redirect"] = ""
	if len(redirect) > 0 {
		out["redirect"] = redirect[0]
	}

	if c.IsAjax() {
		c.jsonResult(out)
	} else {
		for k, v := range out {
			c.Data[k] = v
		}
		c.display("error/message")
		c.Render()
		c.StopRun()
	}
}

//获取用户IP地址
func (c *BaseController) getClientIp() string {
	if p := c.Ctx.Input.Proxy(); len(p) > 0 {
		return p[0]
	}
	return c.Ctx.Input.IP()
}

// 功能菜单
func (c *BaseController) getMenuList() []Menu {
	var menuList []Menu
	allMenu := make([]Menu, 0)
	content, err := ioutil.ReadFile("conf/menu.json")
	if err == nil {
		err := json.Unmarshal(content, &allMenu)
		if err != nil {
			beego.Error(err.Error())
		}
	}
	menuList = make([]Menu, 0)
	for _, menu := range allMenu {
		subs := make([]SubMenu, 0)
		for _, sub := range menu.Submenu {
			route := strings.Split(sub.Route, ".")
			if c.auth.HasAccessPerm(route[0], route[1]) {
				subs = append(subs, sub)
			}
		}
		if len(subs) > 0 {
			menu.Submenu = subs
			menuList = append(menuList, menu)
		}
	}
	//menuList = allMenu
	return menuList
}

// 输出json
func (c *BaseController) jsonResult(out interface{}) {
	c.Data["json"] = out
	c.ServeJSON()
	c.StopRun()
}

// 错误检查
func (c *BaseController) checkError(err error) {
	if err != nil {
		c.showMsg(err.Error(), MSG_ERR)
	}
}
