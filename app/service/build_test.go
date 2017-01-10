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

func TestBuild(t *testing.T) {

}
