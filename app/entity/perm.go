package entity

type Perm struct {
	Module string `orm:"size(20)"`
	Action string `orm:"size(20)"`
	Key    string `orm:"-"` // Module.Action
}

func (p *Perm) TableUnique() [][]string {
	return [][]string{
		[]string{"Module", "Action"},
	}
}
