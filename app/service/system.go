package service

import (
	"../entity"
)

type systemService struct{}

func (this *systemService) GetPermList() map[string][]entity.Perm {
	var list []entity.Perm
	o.Raw("SELECT * FROM " + tableName("perm")).QueryRows(&list)

	result := make(map[string][]entity.Perm)
	for _, v := range list {
		v.Key = v.Module + "." + v.Action
		if _, ok := result[v.Module]; !ok {
			result[v.Module] = make([]entity.Perm, 0)
		}
		result[v.Module] = append(result[v.Module], v)
	}
	return result
}
