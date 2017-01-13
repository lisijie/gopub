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
    "sync"
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

type deployService struct {
    runningTasks map[int]*DeployTask
    lock         sync.Mutex
}

func (m *deployService) DeployTask(taskId int) error {
    if m.IsRunning(taskId) {
        return fmt.Errorf("正在发布中")
    }
    m.lock.Lock()
    defer m.lock.Unlock()
    task, err := TaskService.GetTask(taskId)
    if err != nil {
        return err
    }
    dt := NewDeployTask(task)
    m.runningTasks[task.Id] = dt
    go func() {
        dt.Deploy()
        delete(m.runningTasks, task.Id)
    }()
    return nil
}

func (m *deployService) IsRunning(taskId int) bool {
    m.lock.Lock()
    defer m.lock.Unlock()
    _, ok := m.runningTasks[taskId]
    return ok
}

// 中断发布
func (m *deployService) Abort(taskId int) error {
    m.lock.Lock()
    defer m.lock.Unlock()
    if v, ok := m.runningTasks[taskId]; ok {
        return v.Abort()
    }
    return fmt.Errorf("任务未发布或已发布完成")
}

func (m *deployService) GetMessage(taskId int) (string, error) {
    if !m.IsRunning(taskId) {
        task, err := TaskService.GetTask(taskId)
        if err != nil {
            return "", err
        }
        return task.PubLog, nil
    }
    return m.runningTasks[taskId].Message(), nil
}

type DeployTask struct {
    task    *entity.Task
    logFile string
    message bytes.Buffer
}

func NewDeployTask(task *entity.Task) *DeployTask {
    return &DeployTask{task:task, logFile:Setting.GetTaskPath(task.Id) + "/deploy.log"}
}

// 中断部署
func (s *DeployTask) Abort() error {
    return nil
}

func (s *DeployTask) Message() string {
    return s.message.String()
}

// 执行部署任务
func (s *DeployTask) Deploy() {
    trace("开始执行部署任务, ID:", s.task.Id)
    s.task.PubStatus = 1
    s.task.ErrorMsg = ""
    TaskService.UpdateTask(s.task, "PubStatus", "ErrorMsg")

    // 1. 发布到跳板机
    s.writeLog("开始上传更新包到中转服务器...")
    err := s.syncToAgent()
    if err != nil {
        s.writeLog(err)
        s.task.ErrorMsg = fmt.Sprintf("发布到跳板机失败：%v", err)
        s.task.PubStatus = -2
        TaskService.UpdateTask(s.task, "PubStatus", "ErrorMsg", "PubLog")
        return
    }

    // 2. 发布到目标服务器
    s.task.ErrorMsg = ""
    s.task.PubStatus = 2
    s.writeLog("登录中转服务器执行发布脚本...")
    err = s.syncToServer()
    if err != nil {
        s.writeLog(err)
        s.task.PubStatus = -3
        s.task.ErrorMsg = err.Error()
        TaskService.UpdateTask(s.task, "PubStatus", "ErrorMsg", "PubLog")
        return
    }

    // 更新项目的最后发步版本
    project, _ := ProjectService.GetProject(s.task.ProjectId)
    project.Version = s.task.EndVer
    project.VersionTime = time.Now()
    ProjectService.UpdateProject(project, "Version", "VersionTime")

    // 3. 发送邮件
    env, _ := EnvService.GetEnv(s.task.PubEnvId)
    if env.SendMail > 0 {
        s.writeLog("发送邮件通知相关人员...")
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
                s.writeLog(err)
            }
        }
    }

    s.writeLog("部署完成.")
    s.task.PubTime = time.Now()
    s.task.PubStatus = 3
    s.task.ErrorMsg = ""
    TaskService.UpdateTask(s.task, "PubTime", "PubLog", "PubStatus", "ErrorMsg")
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
    s.writeLog("连接跳板机: ", addr, ", 用户: ", agentServer.SshUser)

    // 上传更新包
    srcFile = s.task.FilePath
    dstFile = filepath.Join(agentTaskDir, filepath.Base(s.task.FilePath))
    err = server.CopyFile(srcFile, dstFile)
    s.writeLog("上传更新包: ", srcFile, " ==> ", dstFile, ", 错误: ", err)
    if err != nil {
        return err
    }

    // 上传更新脚本
    srcFile = s.task.ScriptPath
    dstFile = filepath.Join(agentTaskDir, filepath.Base(srcFile))
    err = server.CopyFile(srcFile, dstFile)
    s.writeLog("上传更新脚本: ", srcFile, " ==> ", dstFile, ", 错误: ", err)
    if err != nil {
        return err
    }

    return nil
}

// 发布到线上服务器
// 登录到中转服务器上执行发布脚本
func (s *DeployTask) syncToServer() error {
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
    s.writeLog("连接跳板机: ", addr, ", 用户: ", agentServer.SshUser, ", Key: ", agentServer.SshKey)
    // 执行发布脚本
    scriptFile := filepath.Join(agentTaskDir, filepath.Base(s.task.ScriptPath))
    out := make(chan string)
    go func() {
        for line := range out {
            s.writeLog("> " + line)
        }
    }()
    s.writeLog("在跳板机执行发布脚本: ", scriptFile)
    err = server.RunCmdPipe("/bin/bash " + scriptFile, out)
    return err
}

func (s *DeployTask) writeLog(v ...interface{}) error {
    f, err := os.OpenFile(s.logFile, os.O_CREATE | os.O_RDWR | os.O_APPEND, 0666)
    if err != nil {
        return err
    }
    defer f.Close()
    ts := time.Now().Format("2006-01-02 15:04:05") + " "
    fmt.Fprint(f, ts)
    fmt.Fprintln(f, v...)
    s.message.WriteString(ts + fmt.Sprint(v...) + "\n")
    s.task.PubLog = s.message.String()
    trace(v...)
    return nil
}
