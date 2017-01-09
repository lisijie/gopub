package controllers

import (
	"github.com/astaxie/beego"
	"github.com/lisijie/gopub/app/entity"
	"github.com/lisijie/gopub/app/service"
)

type MailTplController struct {
	BaseController
}

// 模板列表
func (c *MailTplController) List() {
	list, _ := service.MailService.GetMailTplList()
	c.Data["pageTitle"] = "邮件模板"
	c.Data["list"] = list
	c.display()
}

// 添加模板
func (c *MailTplController) Add() {
	if c.isPost() {
		name := c.GetString("name")
		subject := c.GetString("subject")
		content := c.GetString("content")
		mailTo := c.GetString("mail_to")
		mailCc := c.GetString("mail_cc")

		if name == "" || subject == "" || content == "" {
			c.showMsg("模板名称、邮件主题、邮件内容不能为空", MSG_ERR)
		}

		tpl := new(entity.MailTpl)
		tpl.UserId = c.auth.GetUserId()
		tpl.Name = name
		tpl.Subject = subject
		tpl.Content = content
		tpl.MailTo = mailTo
		tpl.MailCc = mailCc
		err := service.MailService.AddMailTpl(tpl)
		c.checkError(err)

		c.redirect(beego.URLFor("MailTplController.List"))
	}

	c.Data["pageTitle"] = "添加模板"
	c.display()
}

// 编辑模板
func (c *MailTplController) Edit() {
	id, _ := c.GetInt("id")
	tpl, err := service.MailService.GetMailTpl(id)
	c.checkError(err)

	if c.isPost() {
		name := c.GetString("name")
		subject := c.GetString("subject")
		content := c.GetString("content")
		mailTo := c.GetString("mail_to")
		mailCc := c.GetString("mail_cc")
		if name == "" || subject == "" || content == "" {
			c.showMsg("模板名称、邮件主题、邮件内容不能为空", MSG_ERR)
		}

		tpl.Name = name
		tpl.Subject = subject
		tpl.Content = content
		tpl.MailTo = mailTo
		tpl.MailCc = mailCc
		err := service.MailService.SaveMailTpl(tpl)
		c.checkError(err)

		c.redirect(beego.URLFor("MailTplController.List"))
	}

	c.Data["pageTitle"] = "修改模板"
	c.Data["tpl"] = tpl
	c.display()
}

// 删除模板
func (c *MailTplController) Del() {
	id, _ := c.GetInt("id")

	err := service.MailService.DelMailTpl(id)
	c.checkError(err)

	c.redirect(beego.URLFor("MailTplController.List"))
}
