package service

import (
    "github.com/lisijie/gopub/app/entity"
    "github.com/lisijie/gopub/app/libs/repository"
    "os"
)

type projectService struct{}

// 表名
func (s projectService) table() string {
    return tableName("project")
}

// 获取一个项目信息
func (s projectService) GetProject(id int) (*entity.Project, error) {
    project := &entity.Project{}
    project.Id = id
    if err := o.Read(project); err != nil {
        return nil, err
    }
    return project, nil
}

// 获取所有项目
func (s projectService) GetAllProject() ([]entity.Project, error) {
    return s.GetList(1, -1)
}

// 获取项目列表
func (s projectService) GetList(page, pageSize int) ([]entity.Project, error) {
    var list []entity.Project
    offset := 0
    if pageSize == -1 {
        pageSize = 100000
    } else {
        offset = (page - 1) * pageSize
        if offset < 0 {
            offset = 0
        }
    }

    _, err := o.QueryTable(s.table()).Offset(offset).Limit(pageSize).All(&list)
    return list, err
}

// 获取项目总数
func (s projectService) GetTotal() (int64, error) {
    return o.QueryTable(s.table()).Count()
}

// 添加项目
func (s projectService) AddProject(project *entity.Project) error {
    _, err := o.Insert(project)
    return err
}

// 更新项目信息
func (s projectService) UpdateProject(project *entity.Project, fields ...string) error {
    _, err := o.Update(project, fields...)
    return err
}

// 删除一个项目
func (s projectService) DeleteProject(projectId int) error {
    project, err := s.GetProject(projectId)
    if err != nil {
        return err
    }
    // 删除目录
    path := GetProjectPath(project.Domain)
    os.RemoveAll(path)
    // 环境配置
    if envList, err := EnvService.GetEnvListByProjectId(project.Id); err != nil {
        for _, env := range envList {
            EnvService.DeleteEnv(env.Id)
        }
    }
    // 删除任务
    TaskService.DeleteByProjectId(project.Id)
    // 删除项目
    o.Delete(project)
    return nil
}

// 克隆某个项目的仓库
func (s projectService) CloneRepo(projectId int) error {
    project, err := s.GetProject(projectId)
    if err != nil {
        return err
    }
    repo, _ := s.GetRepository(project.Id)
    err = repo.Clone()
    if err != nil {
        project.Status = -1
        project.ErrorMsg = err.Error()
    } else {
        project.Status = 1
    }
    ProjectService.UpdateProject(project, "Status", "ErrorMsg")
    return err
}

func (s projectService) GetRepository(projectId int) (repository.Repository, error) {
    project, err := s.GetProject(projectId)
    if err != nil {
        return nil, err
    }
    repo := repository.NewRepository(project.RepoType, &repository.Config{
        ClonePath: GetProjectPath(project.Domain),
        RemoteUrl: project.RepoUrl,
        Username: project.RepoUser,
        Password: project.RepoPassword,
    })
    return repo, err
}
