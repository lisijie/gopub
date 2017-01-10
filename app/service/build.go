package service

import (
    "github.com/lisijie/gopub/app/entity"
    "fmt"
    "strings"
    "path/filepath"
    "os"
    "github.com/lisijie/gopub/app/libs/utils"
)

type buildService struct{}

func (s buildService) BuildTask(task *entity.Task) error {
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
    task.FilePath = filename
    scriptFile, err := s.CreateScript(task)
    if err != nil {
        os.Remove(filename)
        return fmt.Errorf("生成发布脚本失败: %v", err)
    }
    task.ScriptPath = scriptFile
    TaskService.UpdateTask(task, "FilePath", "ScriptPath")
    return nil
}

func (s buildService) CreateScript(task *entity.Task) (string, error) {
    projectInfo, err := ProjectService.GetProject(task.ProjectId)
    if err != nil {
        return "", err
    }
    agentServer, err := ServerService.GetServer(projectInfo.AgentId)
    if err != nil {
        return "", err
    }
    envInfo, err := EnvService.GetEnv(task.PubEnvId)
    if err != nil {
        return "", err
    }
    if envInfo.ServerCount < 1 {
        return "", fmt.Errorf("服务器列表为空")
    }
    scriptFilePath := fmt.Sprintf("%s/task-%d/publish.sh", GetTasksBasePath(), task.Id)
    agentTaskDir := fmt.Sprintf("%s/%s/tasks/task-%d", agentServer.WorkDir, projectInfo.Domain, task.Id)
    agentWwwDir := agentServer.WorkDir + "/" + projectInfo.Domain + "/www_root" // 跳板机的项目目录
    agentBackupDir := agentTaskDir + "/backup"                             // 跳板机的备份目录
    agentTarFile := agentTaskDir + "/" + filepath.Base(task.FilePath) // 跳板机的更新包路径
    serverUser := envInfo.SshUser                                              // 服务器登录帐号
    serverPort := envInfo.SshPort                                              // 服务器登录端口
    serverKey := envInfo.SshKey                                                // 服务器私钥
    serverWwwDir := strings.TrimRight(envInfo.PubDir, "/")                     // 服务器web目录

    // 同步忽略文件列表
    ignoreListCmd := ""
    for _, v := range strings.Split(projectInfo.IgnoreList, "\n") {
        v = strings.TrimSpace(v)
        if v != "" {
            ignoreListCmd = ignoreListCmd + " --exclude=" + v
        }
    }

    ipList := make([]string, 0, len(envInfo.ServerList))
    for _, v := range envInfo.ServerList {
        ipList = append(ipList, v.Ip)
    }

    f, err := os.Create(scriptFilePath)
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

    if projectInfo.CreateVerfile > 0 {
        f.WriteString("echo '创建版本号文件'\n")
        f.WriteString("echo '" + task.EndVer + "' > " + filepath.Join(agentWwwDir, projectInfo.VerfilePath, "version.txt") + "\n")
        f.WriteString("echo `date '+%Y-%m-%d %H:%M:%S'` > " + filepath.Join(agentWwwDir, projectInfo.VerfilePath, "release.txt") + "\n\n")
    }

    f.WriteString("echo '清理不需要的文件'\n")
    f.WriteString("find " + agentWwwDir + " -type f -name \"*.swp\" -delete\n")
    f.WriteString("find " + agentWwwDir + " -type f -name \"*.swo\" -delete\n\n")

    if projectInfo.BeforeShell != "" {
        f.WriteString("echo '在跳板机执行同步之前操作'\n")
        f.WriteString(projectInfo.BeforeShell)
        f.WriteString("\n")
    }

    for _, ip := range ipList {
        if envInfo.BeforeShell != "" {
            f.WriteString("echo 'SSH登录到 " + ip + " 执行同步前脚本'\n")
            f.WriteString("ssh -o StrictHostKeyChecking=no -p " + serverPort + " -i " + serverKey + " " + serverUser + "@" + ip + " '" + envInfo.BeforeShell + "'\n\n")
        }

        f.WriteString("echo '同步文件到 " + ip + "'\n")
        f.WriteString("rsync -aqc -e 'ssh -o StrictHostKeyChecking=no -p " + serverPort + " -i " + serverKey + "' " + agentWwwDir + "/ " + serverUser + "@" + ip + ":" + serverWwwDir + "/ " + ignoreListCmd + "\n\n")

        if envInfo.AfterShell != "" {
            f.WriteString("echo 'SSH登录到 " + ip + " 执行同步后脚本'\n")
            f.WriteString("ssh -o StrictHostKeyChecking=no -p " + serverPort + " -i " + serverKey + " " + serverUser + "@" + ip + " '" + envInfo.AfterShell + "'\n")
        }
    }

    if projectInfo.AfterShell != "" {
        f.WriteString("echo '在跳板机执行同步之后操作'\n")
        f.WriteString(projectInfo.AfterShell)
        f.WriteString("\n")
    }

    f.WriteString("echo '发布完成，感谢使用！'\n")
    f.Sync()

    return scriptFilePath, nil
}
