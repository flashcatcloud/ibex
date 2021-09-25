package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/toolkits/pkg/cache"
	"github.com/toolkits/pkg/str"
	"gorm.io/gorm"
)

type TaskMeta struct {
	Id        int64     `json:"id" gorm:"primaryKey"`
	Title     string    `json:"title"`
	Account   string    `json:"account"`
	Batch     int       `json:"batch"`
	Tolerance int       `json:"tolerance"`
	Timeout   int       `json:"timeout"`
	Pause     string    `json:"pause"`
	Script    string    `json:"script"`
	Args      string    `json:"args"`
	Creator   string    `json:"creator"`
	Created   time.Time `json:"created" gorm:"->"`
	Done      bool      `json:"done" gorm:"-"`
}

func (TaskMeta) TableName() string {
	return "task_meta"
}

func taskMetaCacheKey(k string) string {
	return fmt.Sprintf("/cache/task/meta/%s", k)
}

func TaskMetaGet(where string, args ...interface{}) (*TaskMeta, error) {
	var arr []*TaskMeta
	err := DB().Where(where, args...).Find(&arr).Error
	if err != nil {
		return nil, err
	}

	if len(arr) == 0 {
		return nil, nil
	}

	return arr[0], nil
}

// TaskMetaGet 根据ID获取任务元信息，会用到内存缓存
func TaskMetaGetByID(id interface{}) (*TaskMeta, error) {
	var obj TaskMeta
	if err := cache.Get(taskMetaCacheKey(fmt.Sprint(id)), &obj); err == nil {
		return &obj, nil
	}

	meta, err := TaskMetaGet("id=?", id)
	if err != nil {
		return nil, err
	}

	if meta == nil {
		return nil, nil
	}

	cache.Set(taskMetaCacheKey(fmt.Sprint(id)), *meta, cache.DEFAULT)

	return meta, nil
}

func (m *TaskMeta) CleanFields() error {
	if m.Batch < 0 {
		return fmt.Errorf("arg(batch) should be nonnegative")
	}

	if m.Tolerance < 0 {
		return fmt.Errorf("arg(tolerance) should be nonnegative")
	}

	if m.Timeout < 0 {
		return fmt.Errorf("arg(timeout) should be nonnegative")
	}

	if m.Timeout > 3600*24 {
		return fmt.Errorf("arg(timeout) longer than one day")
	}

	if m.Timeout == 0 {
		m.Timeout = 30
	}

	m.Pause = strings.Replace(m.Pause, "，", ",", -1)
	m.Pause = strings.Replace(m.Pause, " ", "", -1)
	m.Args = strings.Replace(m.Args, "，", ",", -1)

	if m.Title == "" {
		return fmt.Errorf("arg(title) is required")
	}

	if str.Dangerous(m.Title) {
		return fmt.Errorf("arg(title) is dangerous")
	}

	if m.Script == "" {
		return fmt.Errorf("arg(script) is required")
	}

	if str.Dangerous(m.Args) {
		return fmt.Errorf("arg(args) is dangerous")
	}

	if str.Dangerous(m.Pause) {
		return fmt.Errorf("arg(pause) is dangerous")
	}

	return nil
}

func (m *TaskMeta) HandleFH(fh string) {
	i := strings.Index(m.Title, " FH: ")
	if i > 0 {
		m.Title = m.Title[:i]
	}
	m.Title = m.Title + " FH: " + fh
}

func (m *TaskMeta) Save(hosts []string, action string) error {
	if err := m.CleanFields(); err != nil {
		return err
	}

	m.HandleFH(hosts[0])

	return DB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(m).Error; err != nil {
			return err
		}

		id := m.Id

		if err := tx.Create(&TaskScheduler{Id: id}).Error; err != nil {
			return err
		}

		if err := tx.Create(&TaskAction{Id: id, Action: action, Clock: time.Now().Unix()}).Error; err != nil {
			return err
		}

		for i := 0; i < len(hosts); i++ {
			host := strings.TrimSpace(hosts[i])
			if host == "" {
				continue
			}

			err := tx.Table(tht(id)).Create(map[string]interface{}{
				"id":     id,
				"host":   host,
				"status": "waiting",
			}).Error

			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (m *TaskMeta) Action() (*TaskAction, error) {
	return TaskActionGet("id=?", m.Id)
}

func (m *TaskMeta) Hosts() ([]TaskHost, error) {
	var ret []TaskHost
	err := DB().Table(tht(m.Id)).Where("id=?", m.Id).Select("id", "host", "status").Order("ii").Find(&ret).Error
	return ret, err
}

func (m *TaskMeta) KillHost(host string) error {
	bean, err := TaskHostGet(m.Id, host)
	if err != nil {
		return err
	}

	if bean == nil {
		return fmt.Errorf("no such host")
	}

	if !(bean.Status == "running" || bean.Status == "timeout") {
		return fmt.Errorf("current status cannot kill")
	}

	if err := redoHost(m.Id, host, "kill"); err != nil {
		return err
	}

	return statusSet(m.Id, host, "killing")
}

func (m *TaskMeta) IgnoreHost(host string) error {
	return statusSet(m.Id, host, "ignored")
}

func (m *TaskMeta) RedoHost(host string) error {
	bean, err := TaskHostGet(m.Id, host)
	if err != nil {
		return err
	}

	if bean == nil {
		return fmt.Errorf("no such host")
	}

	if err := redoHost(m.Id, host, "start"); err != nil {
		return err
	}

	return statusSet(m.Id, host, "running")
}

func statusSet(id int64, host, status string) error {
	return DB().Table(tht(id)).Where("id=? and host=?", id, host).Update("status", status).Error
}

func redoHost(id int64, host, action string) error {
	count, err := Count(DB().Model(&TaskHostDoing{}).Where("id=? and host=?", id, host))
	if err != nil {
		return err
	}

	now := time.Now().Unix()
	if count == 0 {
		err = DB().Table("task_host_doing").Create(map[string]interface{}{
			"id":     id,
			"host":   host,
			"clock":  now,
			"action": action,
		}).Error
	} else {
		err = DB().Table("task_host_doing").Where("id=? and host=? and action <> ?", id, host, action).Updates(map[string]interface{}{
			"clock":  now,
			"action": action,
		}).Error
	}
	return err
}

func (m *TaskMeta) HostStrs() ([]string, error) {
	var ret []string
	err := DB().Table(tht(m.Id)).Where("id=?", m.Id).Order("ii").Pluck("host", &ret).Error
	return ret, err
}

func (m *TaskMeta) Stdouts() ([]TaskHost, error) {
	var ret []TaskHost
	err := DB().Table(tht(m.Id)).Where("id=?", m.Id).Select("id", "host", "status", "stdout").Order("ii").Find(&ret).Error
	return ret, err
}

func (m *TaskMeta) Stderrs() ([]TaskHost, error) {
	var ret []TaskHost
	err := DB().Table(tht(m.Id)).Where("id=?", m.Id).Select("id", "host", "status", "stderr").Order("ii").Find(&ret).Error
	return ret, err
}

func TaskMetaTotal(creator, query string, before time.Time) (int64, error) {
	session := DB().Model(&TaskMeta{})

	session = session.Where("created > '" + before.Format("2006-01-02 15:04:05") + "'")

	if creator != "" {
		session = session.Where("creator = ?", creator)
	}

	if query != "" {
		// q1 q2 -q3
		arr := strings.Fields(query)
		for i := 0; i < len(arr); i++ {
			if arr[i] == "" {
				continue
			}
			if strings.HasPrefix(arr[i], "-") {
				q := "%" + arr[i][1:] + "%"
				session = session.Where("title not like ?", q)
			} else {
				q := "%" + arr[i] + "%"
				session = session.Where("title like ?", q)
			}
		}
	}

	return Count(session)
}

func TaskMetaGets(creator, query string, before time.Time, limit, offset int) ([]TaskMeta, error) {
	session := DB().Model(&TaskMeta{}).Order("created desc").Limit(limit).Offset(offset)

	session = session.Where("created > '" + before.Format("2006-01-02 15:04:05") + "'")

	if creator != "" {
		session = session.Where("creator = ?", creator)
	}

	if query != "" {
		// q1 q2 -q3
		arr := strings.Fields(query)
		for i := 0; i < len(arr); i++ {
			if arr[i] == "" {
				continue
			}
			if strings.HasPrefix(arr[i], "-") {
				q := "%" + arr[i][1:] + "%"
				session = session.Where("title not like ?", q)
			} else {
				q := "%" + arr[i] + "%"
				session = session.Where("title like ?", q)
			}
		}
	}

	var objs []TaskMeta
	err := session.Find(&objs).Error
	return objs, err
}
