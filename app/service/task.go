package service

import (
	"fmt"
	"github.com/astaxie/beego/orm"
	"../entity"
	"os"
	"strconv"
	"time"
)

type taskService struct{}

func (this *taskService) table() string {
	return tableName("task")
}

// 删除某个项目下的而所有发布任务
func (this *taskService) DeleteByProjectId(projectId int) error {
	_, err := o.QueryTable(this.table()).Filter("project_id", projectId).Delete()
	return err
}

// 获取一个任务信息
func (this *taskService) GetTask(id int) (*entity.Task, error) {
	task := &entity.Task{}
	task.Id = id
	if err := o.Read(task); err != nil {
		return nil, err
	}
	task.ProjectInfo, _ = ProjectService.GetProject(task.ProjectId)
	return task, nil
}

// 获取任务单列表
func (this *taskService) GetList(page, pageSize int, filters ...interface{}) ([]entity.Task, int64) {
	var (
		list  []entity.Task
		count int64
	)

	offset := (page - 1) * pageSize
	query := o.QueryTable(this.table())

	if len(filters) > 0 {
		l := len(filters)
		for k := 0; k < l; k += 2 {
			field, ok := filters[k].(string)
			if !ok {
				continue
			}
			switch field {
			case "start_date":
				v := fmt.Sprintf("%s 00:00:00", filters[k+1].(string))
				query = query.Filter("create_time__gte", v)
			case "end_date":
				v := fmt.Sprintf("%s 23:59:59", filters[k+1].(string))
				query = query.Filter("create_time__lte", v)
			default:
				v := filters[k+1]
				query = query.Filter(filters[k].(string), v)
			}
		}
	}
	count, _ = query.Count()

	if count > 0 {
		query.OrderBy("-id").Offset(offset).Limit(pageSize).All(&list)
		for k, v := range list {
			if p, err := ProjectService.GetProject(v.ProjectId); err == nil {
				list[k].ProjectInfo = p
			} else {
				list[k].ProjectInfo = new(entity.Project)
			}
		}
	}

	return list, count
}

// 添加任务
func (this *taskService) AddTask(task *entity.Task) error {
	if _, err := EnvService.GetEnv(task.PubEnvId); err != nil {
		return fmt.Errorf("获取环境信息失败: %s", err.Error())
	}
	project, err := ProjectService.GetProject(task.ProjectId)
	if err != nil {
		return fmt.Errorf("获取项目信息失败: %s", err.Error())
	}
	if project.TaskReview > 0 {
		task.ReviewStatus = 0 // 未审批
	} else {
		task.ReviewStatus = 1 // 已审批
	}
	task.PubStatus = 0
	// task.PubTime = time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)
	_, err = o.Insert(task)
	return err
}

// 更新任务信息
func (this *taskService) UpdateTask(task *entity.Task, fields ...string) error {
	_, err := o.Update(task, fields...)
	return err
}

// 删除任务
func (this *taskService) DeleteTask(taskId int) error {
	task, err := this.GetTask(taskId)
	if err != nil {
		return err
	}
	if _, err := o.Delete(task); err != nil {
		return err
	}
	return os.RemoveAll(GetTaskPath(task.Id))
}

// 构建发布包
func (this *taskService) BuildTask(task *entity.Task) {
	err := DeployService.Build(task)
	if err != nil {
		task.BuildStatus = -1
		task.ErrorMsg = err.Error()
	} else {
		task.BuildStatus = 1
		task.ErrorMsg = ""
	}
	this.UpdateTask(task, "BuildStatus", "ErrorMsg")
}

// 任务审批
func (this *taskService) ReviewTask(taskId, userId, status int, message string) error {
	if status != 1 && status != -1 {
		return fmt.Errorf("审批状态无效: %d", status)
	}
	user, err := UserService.GetUser(userId, false)
	if err != nil {
		return err
	}
	task, err := this.GetTask(taskId)
	if err != nil {
		return err
	}
	review := &entity.TaskReview{}
	review.TaskId = task.Id
	review.UserId = user.Id
	review.UserName = user.UserName
	review.Status = status
	review.Message = message
	if _, err := o.Insert(review); err != nil {
		return err
	}

	task.ReviewStatus = status
	return this.UpdateTask(task, "ReviewStatus")
}

// 获取审批信息
func (this *taskService) GetReviewInfo(taskId int) (*entity.TaskReview, error) {
	review := new(entity.TaskReview)
	err := o.QueryTable(tableName("task_review")).Filter("task_id", taskId).OrderBy("-id").Limit(1).One(review)
	return review, err
}

// 获取已发布任务总数
func (this *taskService) GetPubTotal() (int64, error) {
	return o.QueryTable(this.table()).Filter("pub_status", 3).Count()
}

// 发布统计
func (this *taskService) GetPubStat(rangeType string) map[int]int {
	var sql string
	var maps []orm.Params

	switch rangeType {
	case "this_month":
		year, month, _ := time.Now().Date()
		startTime := fmt.Sprintf("%d-%02d-01 00:00:00", year, month)
		endTime := fmt.Sprintf("%d-%02d-31 23:59:59", year, month)
		sql = fmt.Sprintf("SELECT DAY(pub_time) AS date, COUNT(*) AS count FROM %s WHERE pub_time BETWEEN '%s' AND '%s' GROUP BY DAY(pub_time) ORDER BY `date` ASC", this.table(), startTime, endTime)
	case "last_month":
		year, month, _ := time.Now().AddDate(0, -1, 0).Date()
		startTime := fmt.Sprintf("%d-%02d-01 00:00:00", year, month)
		endTime := fmt.Sprintf("%d-%02d-31 23:59:59", year, month)
		sql = fmt.Sprintf("SELECT DAY(pub_time) AS date, COUNT(*) AS count FROM %s WHERE pub_time BETWEEN '%s' AND '%s' GROUP BY DAY(pub_time) ORDER BY `date` ASC", this.table(), startTime, endTime)
	case "this_year":
		year := time.Now().Year()
		startTime := fmt.Sprintf("%d-01-01 00:00:00", year)
		endTime := fmt.Sprintf("%d-12-31 23:59:59", year)
		sql = fmt.Sprintf("SELECT MONTH(pub_time) AS date, COUNT(*) AS count FROM %s WHERE pub_time BETWEEN '%s' AND '%s' GROUP BY MONTH(pub_time) ORDER BY `date` ASC", this.table(), startTime, endTime)
	case "last_year":
		year := time.Now().Year() - 1
		startTime := fmt.Sprintf("%d-01-01 00:00:00", year)
		endTime := fmt.Sprintf("%d-12-31 23:59:59", year)
		sql = fmt.Sprintf("SELECT MONTH(pub_time) AS date, COUNT(*) AS count FROM %s WHERE pub_time BETWEEN '%s' AND '%s' GROUP BY MONTH(pub_time) ORDER BY `date` ASC", this.table(), startTime, endTime)
	}

	num, err := o.Raw(sql).Values(&maps)

	result := make(map[int]int)
	if err == nil && num > 0 {
		for _, v := range maps {
			date, _ := strconv.Atoi(v["date"].(string))
			count, _ := strconv.Atoi(v["count"].(string))
			result[date] = count
		}
	}
	return result
}

func (this *taskService) GetProjectPubStat() []map[string]int {
	var maps []orm.Params
	sql := "SELECT project_id, COUNT(*) AS count FROM " + this.table() + " WHERE pub_status = 3 GROUP BY project_id ORDER BY `count` DESC"
	num, err := o.Raw(sql).Values(&maps)
	result := make([]map[string]int, 0, num)
	if err == nil && num > 0 {
		for _, v := range maps {
			projectId, _ := strconv.Atoi(v["project_id"].(string))
			count, _ := strconv.Atoi(v["count"].(string))
			result = append(result, map[string]int{
				"project_id": projectId,
				"count":      count,
			})
		}
	}
	return result
}
