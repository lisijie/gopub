package service

import (
	"../entity"
)

type mailService struct{}

func (this *mailService) table() string {
	return tableName("mail_tpl")
}

func (this *mailService) AddMailTpl(tpl *entity.MailTpl) error {
	_, err := o.Insert(tpl)
	return err
}

func (this *mailService) DelMailTpl(id int) error {
	_, err := o.QueryTable(this.table()).Filter("id", id).Delete()
	return err
}

func (this *mailService) SaveMailTpl(tpl *entity.MailTpl) error {
	_, err := o.Update(tpl)
	return err
}

func (this *mailService) GetMailTpl(id int) (*entity.MailTpl, error) {
	tpl := &entity.MailTpl{}
	tpl.Id = id
	err := o.Read(tpl)
	return tpl, err
}

// 获取邮件模板列表
func (this *mailService) GetMailTplList() ([]entity.MailTpl, error) {
	var list []entity.MailTpl
	_, err := o.QueryTable(this.table()).OrderBy("-id").All(&list)
	return list, err
}
