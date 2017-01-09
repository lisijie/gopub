package service

import (
    "fmt"
    "github.com/astaxie/beego"
    "github.com/lisijie/gopub/app/entity"
    "github.com/lisijie/gopub/app/libs/ssh"
    "os"
    "path/filepath"
    "strings"
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
 * |	|- www.comprae.com
 * |	|	|- www_root
 * |	|	|- tasks
 * |	|	|	|- task-1
 * |	|	|	|	|- publish.sh
 * |	|	|	|	|- backup
 * |	|	|	|	|- ver1.0-1.1.tar.gz
 *
 *
 */
type DeployJob struct {
    task         *entity.Task    // 任务对象
    project      *entity.Project // 项目对象
    agent        *entity.Server  // 跳板机信息
    env          *entity.Env     // 发布环境
    agentTaskDir string          // 任务在跳板机的目录
    scriptFile   string          // 发布脚本路径
    server       *ssh.ServerConn
}

// 初始化
func (s *DeployJob) init() {
    // 初始化项目信息
    s.project, _ = ProjectService.GetProject(s.task.ProjectId)
    s.agent, _ = ServerService.GetServer(s.project.AgentId)
    // 环境信息
    s.env, _ = EnvService.GetEnv(s.task.PubEnvId)
    // 任务在跳板机的目录
    s.agentTaskDir = fmt.Sprintf("%s/%s/tasks/task-%d", s.agent.WorkDir, s.project.Domain, s.task.Id)
    // 发布脚本路径
    s.scriptFile = fmt.Sprintf("%s/task-%d/publish.sh", GetTasksBasePath(), s.task.Id)
}

// 发布到跳板机
func (s *DeployJob) PubToAgent() error {
    var (
        err error
        srcFile string
        dstFile string
    )
    // 连接到跳板机
    addr := fmt.Sprintf("%s:%d", s.agent.Ip, s.agent.SshPort)
    server := ssh.NewServerConn(addr, s.agent.SshUser, s.agent.SshKey)
    defer server.Close()
    beego.Debug("连接跳板机: ", addr, ", 用户: ", s.agent.SshUser, ", Key: ", s.agent.SshKey)

    // 上传更新包
    srcFile = s.task.Filepath
    dstFile = filepath.Join(s.agentTaskDir, filepath.Base(s.task.Filepath))
    err = server.CopyFile(srcFile, dstFile)
    beego.Debug("上传更新包: ", srcFile, " ==> ", dstFile, ", 错误: ", err)
    if err != nil {
        return err
    }

    // 上传更新脚本
    srcFile = s.scriptFile
    dstFile = filepath.Join(s.agentTaskDir, filepath.Base(s.scriptFile))
    err = server.CopyFile(srcFile, dstFile)
    beego.Debug("上传更新脚本: ", srcFile, " ==> ", dstFile, ", 错误: ", err)
    if err != nil {
        return err
    }

    return nil
}

// 发布到线上服务器
func (s *DeployJob) PubToServer() (string, error) {
    // 连接到跳板机
    addr := fmt.Sprintf("%s:%d", s.agent.Ip, s.agent.SshPort)
    server := ssh.NewServerConn(addr, s.agent.SshUser, s.agent.SshKey)
    defer server.Close()
    beego.Debug("连接跳板机: ", addr, ", 用户: ", s.agent.SshUser, ", Key: ", s.agent.SshKey)
    // 执行发布脚本
    scriptFile := filepath.Join(s.agentTaskDir, filepath.Base(s.scriptFile))
    result, err := server.RunCmd("/bin/bash " + scriptFile)
    beego.Debug("执行发布脚本: ", scriptFile, ", 结果: ", result, ", 错误: ", err)
    return result, err
}

// 创建发布脚本
func (s *DeployJob) CreateScript() (string, error) {
    agentWwwDir := s.agent.WorkDir + "/" + s.project.Domain + "/www_root" // 跳板机的项目目录
    agentBackupDir := s.agentTaskDir + "/backup"                             // 跳板机的备份目录
    agentTarFile := s.agentTaskDir + "/" + filepath.Base(s.task.Filepath) // 跳板机的更新包路径
    serverUser := s.env.SshUser                                              // 服务器登录帐号
    serverPort := s.env.SshPort                                              // 服务器登录端口
    serverKey := s.env.SshKey                                                // 服务器私钥
    serverWwwDir := strings.TrimRight(s.env.PubDir, "/")                     // 服务器web目录

    // 同步忽略文件列表
    ignoreListCmd := ""
    for _, v := range strings.Split(s.project.IgnoreList, "\n") {
        v = strings.TrimSpace(v)
        if v != "" {
            ignoreListCmd = ignoreListCmd + " --exclude=" + v
        }
    }

    // 服务器ip列表
    ipList := s.getServerIpList()
    if len(ipList) < 1 {
        return "", fmt.Errorf("服务器列表为空")
    }

    f, err := os.Create(s.scriptFile)
    if err != nil {
        return "", err
    }
    defer f.Close()

    f.WriteString("#!/bin/bash\n\n")
    f.WriteString("echo '同步第一台web机 " + ipList[0] + " 的代码到本地目录'\n")
    f.WriteString("mkdir -p " + agentWwwDir + "\n")
    f.WriteString("mkdir -p " + agentBackupDir + "\n")
    f.WriteString("rsync -aqc -e 'ssh -o StrictHostKeyChecking=no -p " + serverPort + " -i " + serverKey + "' " + serverUser + "@" + ipList[0] + ":" + serverWwwDir + "/ " + agentWwwDir + "/ " + ignoreListCmd + "\n\n")

    f.WriteString("echo '解压之前，备份 " + agentWwwDir + " 到 " + agentBackupDir + "'\n")
    f.WriteString("mkdir -p " + agentBackupDir + "\n")
    f.WriteString("cp -R " + agentWwwDir + "/* " + agentBackupDir + "\n\n")

    f.WriteString("echo '解压 " + agentTarFile + "'\n")
    f.WriteString("tar -xzf " + agentTarFile + " -C " + agentWwwDir + " " + ignoreListCmd + "\n\n")

    if s.project.CreateVerfile > 0 {
        f.WriteString("echo '创建版本号文件'\n")
        f.WriteString("echo '" + s.task.EndVer + "' > " + filepath.Join(agentWwwDir, s.project.VerfilePath, "version.txt") + "\n")
        f.WriteString("echo `date '+%Y-%m-%d %H:%M:%S'` > " + filepath.Join(agentWwwDir, s.project.VerfilePath, "release.txt") + "\n\n")
    }

    f.WriteString("echo '清理不需要的文件'\n")
    f.WriteString("find " + agentWwwDir + " -type f -name \"*.swp\" -delete\n")
    f.WriteString("find " + agentWwwDir + " -type f -name \"*.swo\" -delete\n\n")

    if s.project.BeforeShell != "" {
        f.WriteString("echo '在跳板机执行同步之前操作'\n")
        f.WriteString(s.project.BeforeShell)
        f.WriteString("\n")
    }

    for _, ip := range ipList {
        if s.env.BeforeShell != "" {
            f.WriteString("echo 'SSH登录到 " + ip + " 执行同步前脚本'\n")
            f.WriteString("ssh -o StrictHostKeyChecking=no -p " + serverPort + " -i " + serverKey + " " + serverUser + "@" + ip + " '" + s.env.BeforeShell + "'\n\n")
        }

        f.WriteString("echo '同步文件到 " + ip + "'\n")
        f.WriteString("rsync -aqc -e 'ssh -o StrictHostKeyChecking=no -p " + serverPort + " -i " + serverKey + "' " + agentWwwDir + "/ " + serverUser + "@" + ip + ":" + serverWwwDir + "/ " + ignoreListCmd + "\n\n")

        if s.env.AfterShell != "" {
            f.WriteString("echo 'SSH登录到 " + ip + " 执行同步后脚本'\n")
            f.WriteString("ssh -o StrictHostKeyChecking=no -p " + serverPort + " -i " + serverKey + " " + serverUser + "@" + ip + " '" + s.env.AfterShell + "'\n")
        }
    }

    if s.project.AfterShell != "" {
        f.WriteString("echo '在跳板机执行同步之后操作'\n")
        f.WriteString(s.project.AfterShell)
        f.WriteString("\n")
    }

    f.WriteString("echo '发布完成，感谢使用！'\n")

    f.Sync()

    return s.scriptFile, nil
}

// 获取服务器ip列表
func (s *DeployJob) getServerIpList() []string {
    ipList := make([]string, 0, len(s.env.ServerList))
    for _, v := range s.env.ServerList {
        ipList = append(ipList, v.Ip)
    }
    return ipList
}

func NewDeployJob(t *entity.Task) *DeployJob {
    job := new(DeployJob)
    job.task = t
    job.init()
    return job
}
