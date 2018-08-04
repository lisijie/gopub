package controllers

import (
	"github.com/astaxie/beego"
	"../entity"
	"../service"
)

type MailTplController struct {
	BaseController
}

// 模板列表
func (this *MailTplController) List() {
	list, _ := service.MailService.GetMailTplList()
	this.Data["pageTitle"] = "邮件模板"
	this.Data["list"] = list
	this.display()
}

// 添加模板
func (this *MailTplController) Add() {
	if this.isPost() {
		name := this.GetString("name")
		subject := this.GetString("subject")
		content := this.GetString("content")
		mailTo := this.GetString("mail_to")
		mailCc := this.GetString("mail_cc")

		if name == "" || subject == "" || content == "" {
			this.showMsg("模板名称、邮件主题、邮件内容不能为空", MSG_ERR)
		}

		tpl := new(entity.MailTpl)
		tpl.UserId = this.auth.GetUserId()
		tpl.Name = name
		tpl.Subject = subject
		tpl.Content = content
		tpl.MailTo = mailTo
		tpl.MailCc = mailCc
		err := service.MailService.AddMailTpl(tpl)
		this.checkError(err)

		this.redirect(beego.URLFor("MailTplController.List"))
	}

	this.Data["pageTitle"] = "添加模板"
	this.display()
}

// 编辑模板
func (this *MailTplController) Edit() {
	id, _ := this.GetInt("id")
	tpl, err := service.MailService.GetMailTpl(id)
	this.checkError(err)

	if this.isPost() {
		name := this.GetString("name")
		subject := this.GetString("subject")
		content := this.GetString("content")
		mailTo := this.GetString("mail_to")
		mailCc := this.GetString("mail_cc")
		if name == "" || subject == "" || content == "" {
			this.showMsg("模板名称、邮件主题、邮件内容不能为空", MSG_ERR)
		}

		tpl.Name = name
		tpl.Subject = subject
		tpl.Content = content
		tpl.MailTo = mailTo
		tpl.MailCc = mailCc
		err := service.MailService.SaveMailTpl(tpl)
		this.checkError(err)

		this.redirect(beego.URLFor("MailTplController.List"))
	}

	this.Data["pageTitle"] = "修改模板"
	this.Data["tpl"] = tpl
	this.display()
}

// 删除模板
func (this *MailTplController) Del() {
	id, _ := this.GetInt("id")

	err := service.MailService.DelMailTpl(id)
	this.checkError(err)

	this.redirect(beego.URLFor("MailTplController.List"))
}
