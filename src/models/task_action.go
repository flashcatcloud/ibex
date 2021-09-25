package models

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type TaskAction struct {
	Id     int64 `gorm:"primaryKey"`
	Action string
	Clock  int64
}

func (TaskAction) TableName() string {
	return "task_action"
}

func TaskActionGet(where string, args ...interface{}) (*TaskAction, error) {
	var obj TaskAction
	ret := DB().Where(where, args...).First(&obj)
	if ret.Error != nil {
		if errors.Is(ret.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, ret.Error
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
