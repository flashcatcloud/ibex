package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type TaskAction struct {
	// id 就是 task id，由调用方赋值，不是自增列。不标 autoIncrement:false 的话 gorm 会
	// 按自增处理，达梦驱动插入前先发 SET IDENTITY_INSERT，而从 MySQL 迁移来的表上没有
	// 自增列，直接报 -2717 表[task_action]不存在IDENTITY列。
	Id     int64  `gorm:"column:id;primaryKey;autoIncrement:false"`
	Action string `gorm:"column:action;size:32;not null"`
	Clock  int64  `gorm:"column:clock;not null;default:0"`
}

func (TaskAction) TableName() string {
	return "task_action"
}

func TaskActionGet(where string, args ...interface{}) (*TaskAction, error) {
	var obj TaskAction
	ret := DB().Where(where, args...).Find(&obj)
	if ret.Error != nil {
		return nil, ret.Error
	}

	if ret.RowsAffected == 0 {
		return nil, nil
	}

	return &obj, nil
}

func TaskActionExistsIds(ids []int64) ([]int64, error) {
	if len(ids) == 0 {
		return ids, nil
	}

	var ret []int64
	err := DB().Model(&TaskAction{}).Where("id in ?", ids).Pluck("id", &ret).Error
	return ret, err
}

func CancelWaitingHosts(id int64) error {
	return DB().Table(tht(id)).Where("id = ? and status = ?", id, "waiting").Update("status", "cancelled").Error
}

func StartTask(id int64) error {
	return DB().Model(&TaskScheduler{}).Where("id = ?", id).Update("scheduler", "").Error
}

func CancelTask(id int64) error {
	return CancelWaitingHosts(id)
}

func KillTask(id int64) error {
	if err := CancelWaitingHosts(id); err != nil {
		return err
	}

	now := time.Now().Unix()

	return DB().Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&TaskHostDoing{}).Where("id = ? and action <> ?", id, "kill").Updates(map[string]interface{}{
			"clock":  now,
			"action": "kill",
		}).Error
		if err != nil {
			return err
		}

		return tx.Table(tht(id)).Where("id = ? and status = ?", id, "running").Update("status", "killing").Error
	})
}

func (a *TaskAction) Update(action string) error {
	if !(action == "start" || action == "cancel" || action == "kill" || action == "pause") {
		return fmt.Errorf("action invalid")
	}

	err := DB().Model(a).Updates(map[string]interface{}{
		"action": action,
		"clock":  time.Now().Unix(),
	}).Error
	if err != nil {
		return err
	}

	if action == "start" {
		return StartTask(a.Id)
	}

	if action == "cancel" {
		return CancelTask(a.Id)
	}

	if action == "kill" {
		return KillTask(a.Id)
	}

	return nil
}

// LongTaskIds two weeks ago
func LongTaskIds() ([]int64, error) {
	clock := time.Now().Unix() - 604800*2
	var ids []int64
	err := DB().Model(&TaskAction{}).Where("clock < ?", clock).Pluck("id", &ids).Error
	return ids, err
}
