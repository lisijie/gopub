package service

import (
    "fmt"
    "github.com/astaxie/beego"
    "github.com/astaxie/beego/orm"
    _ "github.com/go-sql-driver/mysql"
    "github.com/lisijie/gopub/app/entity"
    "net/url"
    "os"
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
)

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
    orm.RunCommand()

    // 创建代码
    os.MkdirAll(GetProjectsBasePath(), 0755)
    os.MkdirAll(GetTasksBasePath(), 0755)

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

// 任务单根目录
func GetTasksBasePath() string {
    return fmt.Sprintf(beego.AppConfig.String("data_dir") + "/tasks")
}

// 所有项目根目录
func GetProjectsBasePath() string {
    return fmt.Sprintf(beego.AppConfig.String("data_dir") + "/projects")
}

// 任务单目录
func GetTaskPath(id int) string {
    return fmt.Sprintf(GetTasksBasePath() + "/task-%d", id)
}

// 某个项目的代码目录
func GetProjectPath(name string) string {
    return GetProjectsBasePath() + "/" + name
}

func DBVersion() string {
    var lists []orm.ParamsList
    o.Raw("SELECT VERSION()").ValuesList(&lists)
    return lists[0][0].(string)
}
