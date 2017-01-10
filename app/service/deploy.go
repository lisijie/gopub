package service

import (
    "fmt"
    "github.com/astaxie/beego"
    "github.com/lisijie/gopub/app/entity"
    "github.com/lisijie/gopub/app/libs/utils"
    "github.com/lisijie/gopub/app/libs/ssh"
    "html"
    "strings"
    "time"
    "path/filepath"
)

/**
 * 发布流程：
 * 1. 打包更新包
 * 2. 生成跳板机同步脚本
 * 3. 把更新包和同步脚本拷贝到跳板机
 * 4. 在跳板机备份旧版本代码
 * 5. 把更新包更新到跳板机项目目录
 * 6. 执行同步脚本，rsync把跳板机的项目目录同步到所有机器
 *
 * 跳板机上面的目录结构如下：
 * |- workspace
 * |	|- www.test.com
 * |	|	|- www_root
 * |	|	|- tasks
 * |	|	|	|- task-1
 * |	|	|	|	|- publish.sh
 * |	|	|	|	|- backup
 * |	|	|	|	|- ver1.0-1.1.tar.gz
 *
 *
 */
type deployService struct{}

// 执行部署任务
func (s deployService) DeployTask(taskId int) error {
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

func (s deployService) doDeploy(task *entity.Task) {
    // 1. 发布到跳板机
    err := s.syncToAgent(task)
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
    ret, err := s.syncToServer(task)
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

// 发布到跳板机
func (s deployService) syncToAgent(task *entity.Task) error {
    var (
        err error
        srcFile string
        dstFile string
    )
    projectInfo, err := ProjectService.GetProject(task.ProjectId)
    if err != nil {
        return err
    }
    agentServer, err := ServerService.GetServer(projectInfo.AgentId)
    if err != nil {
        return err
    }
    agentTaskDir := fmt.Sprintf("%s/%s/tasks/task-%d", agentServer.WorkDir, projectInfo.Domain, task.Id)

    // 连接到跳板机
    addr := fmt.Sprintf("%s:%d", agentServer.Ip, agentServer.SshPort)
    server := ssh.NewServerConn(addr, agentServer.SshUser, agentServer.SshKey)
    defer server.Close()
    beego.Debug("连接跳板机: ", addr, ", 用户: ", agentServer.SshUser, ", Key: ", agentServer.SshKey)

    // 上传更新包
    srcFile = task.FilePath
    dstFile = filepath.Join(agentTaskDir, filepath.Base(task.FilePath))
    err = server.CopyFile(srcFile, dstFile)
    beego.Debug("上传更新包: ", srcFile, " ==> ", dstFile, ", 错误: ", err)
    if err != nil {
        return err
    }

    // 上传更新脚本
    srcFile = task.ScriptPath
    dstFile = filepath.Join(agentTaskDir, filepath.Base(srcFile))
    err = server.CopyFile(srcFile, dstFile)
    beego.Debug("上传更新脚本: ", srcFile, " ==> ", dstFile, ", 错误: ", err)
    if err != nil {
        return err
    }

    return nil
}

// 发布到线上服务器
func (s deployService) syncToServer(task *entity.Task) (string, error) {
    projectInfo, err := ProjectService.GetProject(task.ProjectId)
    if err != nil {
        return "", err
    }
    agentServer, err := ServerService.GetServer(projectInfo.AgentId)
    if err != nil {
        return "", err
    }
    agentTaskDir := fmt.Sprintf("%s/%s/tasks/task-%d", agentServer.WorkDir, projectInfo.Domain, task.Id)
    // 连接到跳板机
    addr := fmt.Sprintf("%s:%d", agentServer.Ip, agentServer.SshPort)
    server := ssh.NewServerConn(addr, agentServer.SshUser, agentServer.SshKey)
    defer server.Close()
    debug("连接跳板机: ", addr, ", 用户: ", agentServer.SshUser, ", Key: ", agentServer.SshKey)
    // 执行发布脚本
    scriptFile := filepath.Join(agentTaskDir, filepath.Base(task.ScriptPath))
    result, err := server.RunCmd("/bin/bash " + scriptFile)
    debug("执行发布脚本: ", scriptFile, ", 结果: ", result, ", 错误: ", err)
    return result, err
}
