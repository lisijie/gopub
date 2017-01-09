package service

import (
    "github.com/lisijie/gopub/app/entity"
    "strings"
    "github.com/astaxie/beego"
    "github.com/lisijie/gopub/app/libs/utils"
    "github.com/lisijie/gomail"
)

type mailService struct{}

func (s *mailService) table() string {
    return tableName("mail_tpl")
}

func (s *mailService) AddMailTpl(tpl *entity.MailTpl) error {
    _, err := o.Insert(tpl)
    return err
}

func (s *mailService) DelMailTpl(id int) error {
    _, err := o.QueryTable(s.table()).Filter("id", id).Delete()
    return err
}

func (s *mailService) SaveMailTpl(tpl *entity.MailTpl) error {
    _, err := o.Update(tpl)
    return err
}

func (s *mailService) GetMailTpl(id int) (*entity.MailTpl, error) {
    tpl := &entity.MailTpl{}
    tpl.Id = id
    err := o.Read(tpl)
    return tpl, err
}

// 获取邮件模板列表
func (s *mailService) GetMailTplList() ([]entity.MailTpl, error) {
    var list []entity.MailTpl
    _, err := o.QueryTable(s.table()).OrderBy("-id").All(&list)
    return list, err
}

func (s *mailService)  SendMail(subject, content string, to, cc []string) error {
    host := beego.AppConfig.String("mail.host")
    port, _ := beego.AppConfig.Int("mail.port")
    username := beego.AppConfig.String("mail.user")
    password := beego.AppConfig.String("mail.password")
    from := beego.AppConfig.String("mail.from")
    if port == 0 {
        port = 25
    }

    toList := make([]string, 0, len(to))
    ccList := make([]string, 0, len(cc))

    for _, v := range to {
        v = strings.TrimSpace(v)
        if utils.IsEmail([]byte(v)) {
            exists := false
            for _, vv := range toList {
                if v == vv {
                    exists = true
                    break
                }
            }
            if !exists {
                toList = append(toList, v)
            }
        }
    }
    for _, v := range cc {
        v = strings.TrimSpace(v)
        if utils.IsEmail([]byte(v)) {
            exists := false
            for _, vv := range ccList {
                if v == vv {
                    exists = true
                    break
                }
            }
            if !exists {
                ccList = append(ccList, v)
            }
        }
    }

    m := gomail.NewMessage()
    m.SetHeader("From", from)
    m.SetHeader("To", toList...)
    if len(ccList) > 0 {
        m.SetHeader("Cc", ccList...)
    }
    m.SetHeader("Subject", subject)
    m.SetBody("text/html", content)

    d := gomail.NewPlainDialer(host, port, username, password)

    return d.DialAndSend(m)
}