package service

import (
    "fmt"
    "github.com/astaxie/beego"
    "github.com/lisijie/gopub/app/entity"
    "github.com/lisijie/gopub/app/libs/utils"
    "html"
    "os"
    "path/filepath"
    "strings"
    "time"
)

type deployService struct{}

// 执行部署任务
func (s *deployService) DeployTask(taskId int) error {
    task, err := TaskService.GetTask(taskId)
    if err != nil {
        return err
    }
    if task.PubStatus > 0 {
        return fmt.Errorf("正在发布或已发布")
    }

    task.PubStatus = 1
    task.ErrorMsg = ""
    TaskService.UpdateTask(task, "PubStatus", "ErrorMsg")

    go s.doDeploy(task)

    return nil
}

func (s *deployService) doDeploy(task *entity.Task) {
    job := NewDeployJob(task)

    // 1. 发布到跳板机
    err := job.PubToAgent()
    if err != nil {
        task.ErrorMsg = fmt.Sprintf("发布到跳板机失败：%v", err)
        task.PubStatus = -2
        TaskService.UpdateTask(task, "PubStatus", "ErrorMsg")
        //s.recordLog("task.publish", fmt.Sprintf("发布到跳板机失败：%v", err))
        return
    }

    // 2. 发布到目标服务器
    task.ErrorMsg = ""
    task.PubStatus = 2
    ret, err := job.PubToServer()
    if err != nil {
        task.PubStatus = -3
        task.ErrorMsg = err.Error()
        TaskService.UpdateTask(task, "PubStatus", "ErrorMsg")
        //s.recordLog("task.publish", fmt.Sprintf("发布到服务器失败：%v", err))
        return
    }
    task.PubTime = time.Now()
    task.PubLog = ret
    task.PubStatus = 3
    task.ErrorMsg = ""
    TaskService.UpdateTask(task, "PubTime", "PubLog", "PubStatus", "ErrorMsg")

    // 更新项目的最后发步版本
    project, _ := ProjectService.GetProject(task.ProjectId)
    project.Version = task.EndVer
    project.VersionTime = time.Now()
    ProjectService.UpdateProject(project, "Version", "VersionTime")

    // 发送邮件
    env, _ := EnvService.GetEnv(task.PubEnvId)
    if env.SendMail > 0 {
        mailTpl, err := MailService.GetMailTpl(env.MailTplId)
        if err == nil {
            replace := make(map[string]string)
            replace["{project}"] = project.Name
            replace["{domain}"] = project.Domain
            if task.StartVer != "" {
                replace["{version}"] = task.StartVer + " - " + task.EndVer
            } else {
                replace["{version}"] = task.EndVer
            }

            replace["{env}"] = env.Name
            replace["{description}"] = utils.Nl2br(html.EscapeString(task.Message))
            replace["{changelogs}"] = utils.Nl2br(html.EscapeString(task.ChangeLogs))
            replace["{changefiles}"] = utils.Nl2br(html.EscapeString(task.ChangeFiles))

            subject := mailTpl.Subject
            content := mailTpl.Content

            for k, v := range replace {
                subject = strings.Replace(subject, k, v, -1)
                content = strings.Replace(content, k, v, -1)
            }

            mailTo := strings.Split(mailTpl.MailTo + "\n" + env.MailTo, "\n")
            mailCc := strings.Split(mailTpl.MailCc + "\n" + env.MailCc, "\n")
            if err := MailService.SendMail(subject, content, mailTo, mailCc); err != nil {
                beego.Error("邮件发送失败：", err)
                //s.recordLog("task.publish", fmt.Sprintf("邮件发送失败：%v", err))
            }
        }
    }
}

func (s *deployService) Build(task *entity.Task) error {
    repo, err := ProjectService.GetRepository(task.ProjectId)
    if err != nil {
        return err
    }
    // 获取版本更新信息
    if task.StartVer != "" {
        logs, err := repo.GetChangeLogs(task.StartVer, task.EndVer)
        if err != nil {
            return fmt.Errorf("获取更新日志失败: %v", err)
        }
        files, err := repo.GetChangeFiles(task.StartVer, task.EndVer)
        if err != nil {
            return fmt.Errorf("获取更新文件列表失败: %v", err)
        }
        task.ChangeLogs = strings.Join(logs, "\n")
        task.ChangeFiles = strings.Join(files, "\n")
        TaskService.UpdateTask(task, "ChangeLogs", "ChangeFiles")
    }

    // 导出目录
    outDir := GetTaskPath(task.Id)
    outDir, _ = filepath.Abs(outDir)
    os.MkdirAll(outDir, 0755)

    // 导出版本号
    var filename string
    if task.StartVer == "" {
        filename = outDir + "/" + task.EndVer + ".tar.gz"
    } else {
        filename = outDir + "/" + task.StartVer + "-" + task.EndVer + ".tar.gz"
    }
    if utils.FileExists(filename) {
        os.Remove(filename)
    }

    // 开始导出
    if task.StartVer == "" {
        err = repo.Export(task.EndVer, filename)
    } else {
        err = repo.ExportDiffFiles(task.StartVer, task.EndVer, filename)
    }
    if err != nil {
        return fmt.Errorf("导出失败(%s): %v", filename, err)
    }
    task.Filepath = filename
    TaskService.UpdateTask(task, "Filepath")

    job := NewDeployJob(task)
    if _, err := job.CreateScript(); err != nil {
        return fmt.Errorf("生成更新脚本失败: %v", err)
    }

    return nil
}
