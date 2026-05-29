package models

import (
	"errors"

	"gorm.io/gorm"
)

// TaskRecord 对应 nightingale center 维护的 task_record 表，只读反查 group_id，不参与写入/迁移。
type TaskRecord struct {
	Id      int64 `gorm:"column:id;primaryKey"`
	GroupId int64 `gorm:"column:group_id"`
}

func (TaskRecord) TableName() string { return "task_record" }

// TaskRecordGroupId 根据 task id 反查所属业务组 id；未找到时返回 gid=0、err=nil。
func TaskRecordGroupId(id int64) (int64, error) {
	var r TaskRecord
	err := DB().Select("id", "group_id").Where("id = ?", id).Take(&r).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	return r.GroupId, err
}
