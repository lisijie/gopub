package repository

import (
    "time"
    "github.com/lisijie/gopub/app/libs/repository/git"
)

type Repository interface {
    // 克隆仓库
    Clone() error
    // 更新代码
    Update() error
    // 获取标签列表
    GetTags() ([]string, error)
    // 获取分支列表
    GetBranches() ([]string, error)
    // 导出整个分支
    Export(branch string, filename string) error
    // 导出两个分支的差异文件
    ExportDiffFiles(ver1 string, ver2 string, filename string) error
    // 获取修改列表
    GetChangeList() (*ChangeList, error)
}

type Config struct {
    RemoteUrl string
    ClonePath string
    Username  string
    Password  string
}

type ChangeList struct {
    Logs  []ChangeLog
    Files []ChangeFile
}

type ChangeLog struct {
    Date time.Time
    Msg  string
}

type ChangeFile struct {
    Filename string
}

func NewRepository(t string, config *Config) Repository {
    var repo Repository
    if t == "GIT" {
        repo = &git.GitRepository{
            Path: config.ClonePath,
            RemoteUrl: config.RemoteUrl,
            Username: config.Username,
            Password: config.Password,
        }
    }
    return repo
}