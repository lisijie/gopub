package service

import (
    "testing"
    "github.com/astaxie/beego"
)

func init() {
    beego.LoadAppConfig("ini", "../../conf/app.conf")
    beego.AppConfig.Set("runmode", "test")
    Init()
}

func TestAdd(t *testing.T) {
    id, err := ActionService.Add("test action", "test actor", "test type", 0, "extra")
    if err != nil {
        t.Error(err)
    }
    o.Raw("DELETE FROM " +tableName("action")+ " WHERE id = ?", id).Exec()
}
