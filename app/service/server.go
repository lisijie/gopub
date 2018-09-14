package service

import (
	"errors"
	"../entity"
)

const (
	SERVER_TYPE_NORMAL = 0 // 普通web服务器
	SERVER_TYPE_AGENT  = 1 // 跳板机
)

type serverService struct{}

// 表名
func (this *serverService) table() string {
	return tableName("server")
}

func (this *serverService) GetTotal(typeId int) (int64, error) {
	return o.QueryTable(this.table()).Filter("TypeId", typeId).Count()
}

// 获取一个服务器信息
func (this *serverService) GetServer(id int, types ...int) (*entity.Server, error) {
	var err error
	server := &entity.Server{}
	server.Id = id
	if len(types) == 0 {
		err = o.Read(server)
	} else {
		err = o.QueryTable(this.table()).Filter("id", id).Filter("type_id", types[0]).One(server)
	}
	return server, err
}

// 根据id列表获取记录
func (this *serverService) GetListByIds(ids []int) ([]entity.Server, error) {
	var list []entity.Server
	if len(ids) == 0 {
		return nil, errors.New("ids不能为空")
	}
	params := make([]interface{}, len(ids))
	for k, v := range ids {
		params[k] = v
	}
	_, err := o.QueryTable(this.table()).Filter("id__in", params...).All(&list)
	return list, err
}

// 获取普通服务器列表
func (this *serverService) GetServerList(page, pageSize int) ([]entity.Server, error) {
	var list []entity.Server
	qs := o.QueryTable(this.table()).Filter("TypeId", SERVER_TYPE_NORMAL)
	if pageSize > 0 {
		qs = qs.Limit(pageSize, (page-1)*pageSize)
	}
	_, err := qs.All(&list)
	return list, err
}

// 获取跳板服务器列表
func (this *serverService) GetAgentList(page, pageSize int) ([]entity.Server, error) {
	var list []entity.Server
	qs := o.QueryTable(this.table()).Filter("TypeId", SERVER_TYPE_AGENT)
	if pageSize > 0 {
		qs = qs.Limit(pageSize, (page-1)*pageSize)
	}
	_, err := qs.All(&list)
	return list, err
}

// 添加服务器
func (this *serverService) AddServer(server *entity.Server) error {
	server.Id = 0
	if o.Read(server, "ip"); server.Id > 0 {
		return errors.New("服务器IP已存在:" + server.Ip)
	}
	_, err := o.Insert(server)
	return err
}

// 修改服务器信息
func (this *serverService) UpdateServer(server *entity.Server, fields ...string) error {
	_, err := o.Update(server, fields...)
	return err
}

// 删除服务器
func (this *serverService) DeleteServer(id int) error {
	_, err := o.QueryTable(this.table()).Filter("id", id).Delete()
	if err != nil {
		return err
	}
	return EnvService.DeleteServer(id)
}
