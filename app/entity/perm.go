package entity

type Perm struct {
    Module string
    Action string
    Key    string // Module.Action
}

func (p *Perm) TableUnique() [][]string {
    return [][]string{
        []string{"Module", "Action"},
    }
}
