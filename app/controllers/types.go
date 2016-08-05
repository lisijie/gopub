package controllers

type Menu struct {
	Name    string
	Route   string
	Icon    string
	Submenu []SubMenu
}

type SubMenu struct {
	Name   string
	Route  string
	Action string
}
