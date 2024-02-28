package migrate

import (
	"fmt"
	"github.com/ulricqin/ibex/src/models"
	"github.com/ulricqin/ibex/src/storage"
)

func Migrate() {
	dts := []interface{}{&models.TaskHostDoing{}, &models.TaskAction{}, &models.TaskMeta{},
		&models.TaskScheduler{}, &models.TaskScheduler{}}
	for id := 0; id < 100; id++ {
		th := new(TaskHost)
		th.Name = fmt.Sprintf("task_host_%d", id)
		dts = append(dts, th)
	}

	if err := storage.DB.AutoMigrate(dts); err != nil {
		panic(err)
	}
}

type TaskHost struct {
	Ii     int64 `gorm:"primaryKey"`
	Id     int64
	Host   string
	Status string
	Stdout string
	Stderr string
	Name   string `gorm:"-"`
}

func (t *TaskHost) TableName() string {
	return t.Name
}
