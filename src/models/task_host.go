package models

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ulricqin/ibex/src/pkg/poster"
	"github.com/ulricqin/ibex/src/server/config"
	"github.com/ulricqin/ibex/src/storage"

	"github.com/toolkits/pkg/logger"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TaskHost struct {
	Id     int64  `json:"id"`
	Host   string `json:"host"`
	Status string `json:"status"`
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
}

func (t *TaskHost) MarshalBinary() ([]byte, error) {
	return json.Marshal(t)
}

func (t *TaskHost) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, t)
}

func (t *TaskHost) Upsert() error {
	return DB().Table(tht(t.Id)).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}, {Name: "host"}},
		DoUpdates: clause.AssignmentColumns([]string{"status", "stdout", "stderr"}),
	}).Create(t).Error
}

func (t *TaskHost) Create() error {
	if config.C.IsCenter {
		return DB().Table(tht(t.Id)).Create(t).Error
	}
	return poster.PostByUrls(config.C.CenterApi, "/ibex/v1/task/host", t)
}

func TaskHostUpserts(lst []TaskHost) (map[string]error, error) {
	if len(lst) == 0 {
		return nil, fmt.Errorf("empty list")
	}

	if !config.C.IsCenter {
		return poster.PostByUrlsWithResp[map[string]error](config.C.CenterApi, "/ibex/v1/task/hosts/upsert", lst)
	}

	errs := make(map[string]error, 0)
	for _, th := range lst {
		if err := th.Upsert(); err != nil {
			errs[fmt.Sprintf("%d:%s", th.Id, th.Host)] = err
		}
	}
	return errs, nil
}

func TaskHostGet(id int64, host string) (*TaskHost, error) {
	var ret []*TaskHost
	err := DB().Table(tht(id)).Where("id=? and host=?", id, host).Find(&ret).Error
	if err != nil {
		return nil, err
	}

	if len(ret) == 0 {
		return nil, nil
	}

	return ret[0], nil
}

func MarkDoneStatus(id, clock int64, host, status, stdout, stderr string, alertTriggered ...bool) error {
	if len(alertTriggered) > 0 && alertTriggered[0] {
		return CacheMarkDone(context.Background(), TaskHost{
			Id:     id,
			Host:   host,
			Status: status,
			Stdout: stdout,
			Stderr: stderr,
		})
	}

	if !config.C.IsCenter {
		return poster.PostByUrls(config.C.CenterApi, "/ibex/v1/mark/done", map[string]interface{}{
			"id":     id,
			"clock":  clock,
			"host":   host,
			"status": status,
			"stdout": stdout,
			"stderr": stderr,
		})
	}

	count, err := DBRecordCount(TaskHostDoing{}.TableName(), "id=? and host=? and clock=?", id, host, clock)
	if err != nil {
		return err
	}

	if count == 0 {
		// 如果是timeout了，后来任务执行完成之后，结果又上来了，stdout和stderr最好还是存库，让用户看到
		count, err = DBRecordCount(tht(id), "id=? and host=? and status=?", id, host, "timeout")
		if err != nil {
			return err
		}

		if count == 1 {
			return DB().Table(tht(id)).Where("id=? and host=?", id, host).Updates(map[string]interface{}{
				"status": status,
				"stdout": stdout,
				"stderr": stderr,
			}).Error
		}
		return nil
	}

	return DB().Transaction(func(tx *gorm.DB) error {
		err = tx.Table(tht(id)).Where("id=? and host=?", id, host).Updates(map[string]interface{}{
			"status": status,
			"stdout": stdout,
			"stderr": stderr,
		}).Error
		if err != nil {
			return err
		}

		if err = tx.Where("id=? and host=?", id, host).Delete(&TaskHostDoing{}).Error; err != nil {
			return err
		}

		return nil
	})
}

func CacheMarkDone(ctx context.Context, host TaskHost) error {
	if err := storage.Cache.Del(ctx, hostDoingCacheKey(host.Id, host.Host)).Err(); err != nil {
		return err
	}
	TaskHostCachePush(host)

	return nil
}

func WaitingHostList(id int64, limit ...int) ([]TaskHost, error) {
	var hosts []TaskHost
	session := DB().Table(tht(id)).Where("id = ? and status = 'waiting'", id).Order("ii")
	if len(limit) > 0 {
		session = session.Limit(limit[0])
	}
	err := session.Find(&hosts).Error
	return hosts, err
}

func WaitingHostCount(id int64) (int64, error) {
	return DBRecordCount(tht(id), "id=? and status='waiting'", id)
}

func UnexpectedHostCount(id int64) (int64, error) {
	return DBRecordCount(tht(id), "id=? and status in ('failed', 'timeout', 'killfailed')", id)
}

func IngStatusHostCount(id int64) (int64, error) {
	return DBRecordCount(tht(id), "id=? and status in ('waiting', 'running', 'killing')", id)
}

func RunWaitingHosts(hosts []TaskHost) error {
	count := len(hosts)
	if count == 0 {
		return nil
	}

	now := time.Now().Unix()

	return DB().Transaction(func(tx *gorm.DB) error {
		for i := 0; i < count; i++ {
			if err := tx.Table(tht(hosts[i].Id)).Where("id=? and host=?", hosts[i].Id, hosts[i].Host).Update("status", "running").Error; err != nil {
				return err
			}
			err := tx.Create(&TaskHostDoing{Id: hosts[i].Id, Host: hosts[i].Host, Clock: now, Action: "start"}).Error
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func TaskHostStatus(id int64) ([]TaskHost, error) {
	var ret []TaskHost
	err := DB().Table(tht(id)).Select("id", "host", "status").Where("id=?", id).Order("ii").Find(&ret).Error
	return ret, err
}

func TaskHostGets(id int64) ([]TaskHost, error) {
	var ret []TaskHost
	err := DB().Table(tht(id)).Where("id=?", id).Order("ii").Find(&ret).Error
	return ret, err
}

var (
	taskHostCache = make([]TaskHost, 128)
	taskHostLock  sync.RWMutex
)

func TaskHostCachePush(th TaskHost) {
	taskHostLock.Lock()
	defer taskHostLock.Unlock()

	taskHostCache = append(taskHostCache, th)
}

func TaskHostCachePopAll() []TaskHost {
	taskHostLock.Lock()
	defer taskHostLock.Unlock()

	all := taskHostCache
	taskHostCache = make([]TaskHost, 0, 128)

	return all
}

func ReportCacheResult() error {
	result := TaskHostCachePopAll()
	dones := make([]TaskHost, 0)
	for _, th := range result {
		// id大于redis初始id，说明是edge与center失联时，本地告警规则触发的自愈脚本生成的id
		// 为了防止不同边缘机房生成的脚本任务id相同，不上报结果至数据库
		if th.Id >= storage.IDINITIAL {
			logger.Infof("task[%s] host[%s] done, result:[%v]", th.Id, th.Host, th)
		} else {
			dones = append(dones, th)
		}
	}

	if len(dones) == 0 {
		return nil
	}

	errs, err := TaskHostUpserts(dones)
	if err != nil {
		return err
	}
	for key, err := range errs {
		logger.Warningf("report task_host_cache[%s] result error: %v", key, err)
	}
	return nil
}
