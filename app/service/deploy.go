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
    "os"
    "bytes"
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
type DeployTask struct {
    task *entity.Task
}

func NewDeployTask(task *entity.Task) *DeployTask {
    return &DeployTask{task:task}
}

// 执行部署任务
func (s *DeployTask) Deploy() error {
    if s.task.PubStatus > 0 {
        return fmt.Errorf("正在发布或已发布")
    }

    s.task.PubStatus = 1
    s.task.ErrorMsg = ""
    TaskService.UpdateTask(s.task, "PubStatus", "ErrorMsg")

    go s.doDeploy()

    return nil
}

func (s *DeployTask) doDeploy() {
    trace("开始执行部署任务, ID:", s.task.Id)

    // 1. 发布到跳板机
    s.WriteLog("开始上传更新包到中转服务器...")
    err := s.syncToAgent()
    if err != nil {
        s.WriteLog(err)
        s.task.ErrorMsg = fmt.Sprintf("发布到跳板机失败：%v", err)
        s.task.PubStatus = -2
        TaskService.UpdateTask(s.task, "PubStatus", "ErrorMsg")
        //s.recordLog("task.publish", fmt.Sprintf("发布到跳板机失败：%v", err))
        return
    }

    // 2. 发布到目标服务器
    s.task.ErrorMsg = ""
    s.task.PubStatus = 2
    s.WriteLog("登录中转服务器执行发布脚本...")
    ret, err := s.syncToServer()
    if err != nil {
        s.WriteLog(err)
        s.task.PubStatus = -3
        s.task.ErrorMsg = err.Error()
        TaskService.UpdateTask(s.task, "PubStatus", "ErrorMsg")
        //s.recordLog("task.publish", fmt.Sprintf("发布到服务器失败：%v", err))
        return
    }
    s.task.PubTime = time.Now()
    s.task.PubLog = ret
    s.task.PubStatus = 3
    s.task.ErrorMsg = ""
    TaskService.UpdateTask(s.task, "PubTime", "PubLog", "PubStatus", "ErrorMsg")

    // 更新项目的最后发步版本
    project, _ := ProjectService.GetProject(s.task.ProjectId)
    project.Version = s.task.EndVer
    project.VersionTime = time.Now()
    ProjectService.UpdateProject(project, "Version", "VersionTime")

    // 3. 发送邮件
    env, _ := EnvService.GetEnv(s.task.PubEnvId)
    if env.SendMail > 0 {
        s.WriteLog("发送邮件通知相关人员...")
        mailTpl, err := MailService.GetMailTpl(env.MailTplId)
        if err == nil {
            replace := make(map[string]string)
            replace["{project}"] = project.Name
            replace["{domain}"] = project.Domain
            if s.task.StartVer != "" {
                replace["{version}"] = s.task.StartVer + " - " + s.task.EndVer
            } else {
                replace["{version}"] = s.task.EndVer
            }

            replace["{env}"] = env.Name
            replace["{description}"] = utils.Nl2br(html.EscapeString(s.task.Message))
            replace["{changelogs}"] = utils.Nl2br(html.EscapeString(s.task.ChangeLogs))
            replace["{changefiles}"] = utils.Nl2br(html.EscapeString(s.task.ChangeFiles))

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
                s.WriteLog(err)
            }
        }
    }

    s.WriteLog("部署完成.")
}

// 发布到中转服务器
// 登录到中转服务器，将发布包和发布脚本拷贝到服务器
// 当发布包很大时，拷贝时间可能会很长，因此发布系统服务器跟中转服务器最好在同一个内部网
func (s *DeployTask) syncToAgent() error {
    var (
        err error
        srcFile string
        dstFile string
    )
    projectInfo, err := ProjectService.GetProject(s.task.ProjectId)
    if err != nil {
        return err
    }
    agentServer, err := ServerService.GetServer(projectInfo.AgentId)
    if err != nil {
        return err
    }
    agentTaskDir := fmt.Sprintf("%s/%s/tasks/task-%d", agentServer.WorkDir, projectInfo.Domain, s.task.Id)

    // 连接到跳板机
    addr := fmt.Sprintf("%s:%d", agentServer.Ip, agentServer.SshPort)
    server := ssh.NewServerConn(&ssh.Config{
        Addr:addr,
        User:agentServer.SshUser,
        Password:agentServer.SshPwd,
        Key:agentServer.SshKey,
    })
    defer server.Close()
    s.WriteLog("连接跳板机: ", addr, ", 用户: ", agentServer.SshUser, ", Key: ", agentServer.SshKey)

    // 上传更新包
    srcFile = s.task.FilePath
    dstFile = filepath.Join(agentTaskDir, filepath.Base(s.task.FilePath))
    err = server.CopyFile(srcFile, dstFile)
    s.WriteLog("上传更新包: ", srcFile, " ==> ", dstFile, ", 错误: ", err)
    if err != nil {
        return err
    }

    // 上传更新脚本
    srcFile = s.task.ScriptPath
    dstFile = filepath.Join(agentTaskDir, filepath.Base(srcFile))
    err = server.CopyFile(srcFile, dstFile)
    s.WriteLog("上传更新脚本: ", srcFile, " ==> ", dstFile, ", 错误: ", err)
    if err != nil {
        return err
    }

    return nil
}

// 发布到线上服务器
// 登录到中转服务器上执行发布脚本
func (s *DeployTask) syncToServer() (string, error) {
    projectInfo, err := ProjectService.GetProject(s.task.ProjectId)
    if err != nil {
        return "", err
    }
    agentServer, err := ServerService.GetServer(projectInfo.AgentId)
    if err != nil {
        return "", err
    }
    agentTaskDir := fmt.Sprintf("%s/%s/tasks/task-%d", agentServer.WorkDir, projectInfo.Domain, s.task.Id)
    // 连接到跳板机
    addr := fmt.Sprintf("%s:%d", agentServer.Ip, agentServer.SshPort)
    server := ssh.NewServerConn(&ssh.Config{
        Addr:addr,
        User:agentServer.SshUser,
        Password:agentServer.SshPwd,
        Key:agentServer.SshKey,
    })
    defer server.Close()
    s.WriteLog("连接跳板机: ", addr, ", 用户: ", agentServer.SshUser, ", Key: ", agentServer.SshKey)
    // 执行发布脚本
    scriptFile := filepath.Join(agentTaskDir, filepath.Base(s.task.ScriptPath))
    var result bytes.Buffer
    out := make(chan string)
    go func() {
        for line := range out {
            result.WriteString(line)
            s.WriteLog("> " + line)
        }
    }()
    s.WriteLog("在跳板机执行发布脚本: ", scriptFile)
    err = server.RunCmdPipe("/bin/bash " + scriptFile, out)
    return result.String(), err
}

func (s *DeployTask) WriteLog(v ...interface{}) error {
    logFile := Setting.GetTaskPath(s.task.Id) + "/deploy.log"
    f, err := os.OpenFile(logFile, os.O_CREATE | os.O_RDWR | os.O_APPEND, 0666)
    if err != nil {
        return err
    }
    defer f.Close()
    fmt.Fprint(f, time.Now().Format("2006-01-02 15:04:05") + " ")
    fmt.Fprintln(f, v...)
    trace(v...)
    return nil
}
