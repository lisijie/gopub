package service

import (
    "fmt"
    "github.com/astaxie/beego"
    "github.com/astaxie/beego/orm"
    _ "github.com/go-sql-driver/mysql"
    "github.com/lisijie/gopub/app/entity"
    "net/url"
    "os"
    "path/filepath"
)

var (
    o orm.Ormer
    tablePrefix string                    // 表前缀
    UserService       *userService       // 用户服务
    RoleService       *roleService       // 角色服务
    EnvService        *envService        // 发布环境服务
    ServerService     *serverService     // 服务器服务
    ProjectService    *projectService    // 项目服务
    MailService       *mailService       // 邮件服务
    TaskService       *taskService       // 任务服务
    DeployService     *deployService     // 部署服务
    SystemService     *systemService
    ActionService     *actionService     // 系统动态
    BuildService      *buildService      // 构建服务
    Setting           *setting           // 系统设置
)

type setting struct {
    DataPath        string
    TaskBasePath    string
    ProjectBasePath string
}

func (s setting) GetTaskPath(id int) string {
    return fmt.Sprintf(s.TaskBasePath + "/task-%d", id)
}

func (s setting) GetProjectPath(name string) string {
    return s.ProjectBasePath + "/" + name
}

func Init() {
    dbHost := beego.AppConfig.String("db.host")
    dbPort := beego.AppConfig.String("db.port")
    dbUser := beego.AppConfig.String("db.user")
    dbPassword := beego.AppConfig.String("db.password")
    dbName := beego.AppConfig.String("db.name")
    timezone := beego.AppConfig.String("db.timezone")
    tablePrefix = beego.AppConfig.String("db.prefix")

    if dbPort == "" {
        dbPort = "3306"
    }
    dsn := dbUser + ":" + dbPassword + "@tcp(" + dbHost + ":" + dbPort + ")/" + dbName + "?charset=utf8"
    if timezone != "" {
        dsn = dsn + "&loc=" + url.QueryEscape(timezone)
    }
    orm.RegisterDataBase("default", "mysql", dsn)

    orm.RegisterModelWithPrefix(tablePrefix,
        new(entity.Action),
        new(entity.Env),
        new(entity.EnvServer),
        new(entity.MailTpl),
        //new(entity.Perm),
        new(entity.Project),
        new(entity.Role),
        //new(entity.RolePerm),
        new(entity.Server),
        new(entity.Task),
        new(entity.TaskReview),
        new(entity.User),
        //new(entity.UserRole),
    )

    if beego.AppConfig.String("runmode") == "dev" {
        orm.Debug = true
    }

    o = orm.NewOrm()

    // 初始化配置
    initSetting()
    // 初始化服务对象
    initService()
}

func initService() {
    UserService = &userService{}
    RoleService = &roleService{}
    EnvService = &envService{}
    ServerService = &serverService{}
    ProjectService = &projectService{}
    MailService = &mailService{}
    TaskService = &taskService{}
    DeployService = &deployService{}
    SystemService = &systemService{}
    ActionService = &actionService{}
    BuildService = &buildService{}
}

func initSetting() {
    Setting = &setting{}
    Setting.DataPath = beego.AppConfig.String("data_dir")
    if Setting.DataPath == "" {
        p, _ := filepath.Abs(filepath.Dir(os.Args[0]))
        Setting.DataPath = filepath.Join(p, "data")
    }
    Setting.TaskBasePath = filepath.Join(Setting.DataPath, "tasks")
    Setting.ProjectBasePath = filepath.Join(Setting.DataPath, "projects")
    os.MkdirAll(Setting.ProjectBasePath, 0755)
    os.MkdirAll(Setting.TaskBasePath, 0755)
}

// 返回真实表名
func tableName(name string) string {
    return tablePrefix + name
}

func debug(v ...interface{}) {
    beego.Debug(v...)
}

func trace(v ...interface{}) {
    beego.Trace(v ...)
}

func DBVersion() string {
    var lists []orm.ParamsList
    o.Raw("SELECT VERSION()").ValuesList(&lists)
    return lists[0][0].(string)
}