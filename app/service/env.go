package service

import (
	"github.com/astaxie/beego/orm"
	"github.com/lisijie/gopub/app/entity"
)

type envService struct{}

func (this *envService) table() string {
	return tableName("env")
}
func (this *envService) serverTable() string {
	return tableName("env_server")
}

// 获取一个发布环境信息
func (this *envService) GetEnv(id int) (*entity.Env, error) {
	env := &entity.Env{}
	env.Id = id
	err := o.Read(env)
	if err == nil {
		env.ServerList, _ = this.GetEnvServers(env.Id)
	}
	return env, err
}

// 获取某个项目的发布环境列表
func (this *envService) GetEnvListByProjectId(projectId int) ([]entity.Env, error) {
	var list []entity.Env
	_, err := o.QueryTable(this.table()).Filter("project_id", projectId).All(&list)
	for _, env := range list {
		env.ServerList, _ = this.GetEnvServers(env.Id)
	}
	return list, err
}

// 根据服务器id发布环境列表
func (this *envService) GetEnvListByServerId(serverId int) ([]entity.Env, error) {
	var (
		servList []entity.EnvServer
		envList  []entity.Env
	)
	o.QueryTable(this.serverTable()).Filter("server_id", serverId).All(&servList)
	envIds := make([]int, 0, len(servList))
	for _, serv := range servList {
		envIds = append(envIds, serv.EnvId)
	}
	envList = make([]entity.Env, 0)
	if len(envIds) > 0 {
		if _, err := o.QueryTable(this.table()).Filter("id__in", envIds).All(&envList); err != nil {
			return envList, err
		}
	}
	return envList, nil
}

// 获取某个发布环境的服务器列表
func (this *envService) GetEnvServers(envId int) ([]entity.Server, error) {
	var (
		list []entity.EnvServer
	)
	_, err := o.QueryTable(this.serverTable()).Filter("env_id", envId).All(&list)
	if err != nil {
		return nil, err
	}
	servIds := make([]int, 0, len(list))
	for _, v := range list {
		servIds = append(servIds, v.ServerId)
	}

	return ServerService.GetListByIds(servIds)
}

// 新增发布环境
func (this *envService) AddEnv(env *entity.Env) error {
	env.ServerCount = len(env.ServerList)
	if _, err := o.Insert(env); err != nil {
		return err
	}
	for _, sv := range env.ServerList {
		es := new(entity.EnvServer)
		es.ProjectId = env.ProjectId
		es.EnvId = env.Id
		es.ServerId = sv.Id
		o.Insert(es)
	}
	return nil
}

// 保存环境配置
func (this *envService) SaveEnv(env *entity.Env) error {
	env.ServerCount = len(env.ServerList)
	if _, err := o.Update(env); err != nil {
		return err
	}
	o.QueryTable(this.serverTable()).Filter("env_id", env.Id).Delete()
	for _, sv := range env.ServerList {
		es := new(entity.EnvServer)
		es.ProjectId = env.ProjectId
		es.EnvId = env.Id
		es.ServerId = sv.Id
		o.Insert(es)
	}
	return nil
}

// 删除发布环境
func (this *envService) DeleteEnv(id int) error {
	o.QueryTable(this.table()).Filter("id", id).Delete()
	o.QueryTable(this.serverTable()).Filter("env_id", id).Delete()
	return nil
}

// 删除服务器
func (this *envService) DeleteServer(serverId int) error {
	var envServers []entity.EnvServer
	o.QueryTable(this.serverTable()).Filter("server_id", serverId).All(&envServers)
	if len(envServers) < 1 {
		return nil
	}
	envIds := make([]int, 0, len(envServers))
	for _, v := range envServers {
		envIds = append(envIds, v.EnvId)
	}
	o.QueryTable(this.serverTable()).Filter("server_id", serverId).Delete()
	o.QueryTable(this.table()).Filter("id__in", envIds).Update(orm.Params{
		"server_count": orm.ColValue(orm.ColMinus, 1),
	})
	return nil
}
