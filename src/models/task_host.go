package models

import (
	"time"

	"gorm.io/gorm"
)

type TaskHost struct {
	Id     int64  `json:"id" gorm:"primaryKey"`
	Host   string `json:"host"`
	Status string `json:"status"`
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
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

func MarkDoneStatus(id, clock int64, host, status, stdout, stderr string) error {
	count, err := DoingHostCount("id=? and host=? and clock=?", id, host, clock)
	if err != nil {
		return err
	}

	if count == 0 {
		// 如果是timeout了，后来任务执行完成之后，结果又上来了，stdout和stderr最好还是存库，让用户看到
		err = DB().Table(tht(id)).Where("id=? and host=? and status=?", id, host, "timeout").Count(&count).Error
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
		err = DB().Table(tht(id)).Where("id=? and host=?", id, host).Updates(map[string]interface{}{
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
	return Count(DB().Table(tht(id)).Where("id=? and status='waiting'", id))
}

func UnexpectedHostCount(id int64) (int64, error) {
	return Count(DB().Table(tht(id)).Where("id=? and status in ('failed', 'timeout', 'killfailed')", id))
}

func IngStatusHostCount(id int64) (int64, error) {
	return Count(DB().Table(tht(id)).Where("id=? and status in ('waiting', 'running', 'killing')", id))
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
