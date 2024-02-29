package migrate

import (
	"fmt"
	"github.com/ulricqin/ibex/src/storage"
)

func Migrate() {
	//dts := []interface{}{&models.TaskHostDoing{}, &models.TaskAction{}, &models.TaskMeta{},
	//	&models.TaskScheduler{}, &models.TaskScheduler{}}
	dts := make([]interface{}, 0)
	for id := 0; id < 100; id++ {
		th := new(TaskHost)
		tname := fmt.Sprintf("task_host_%d", id)
		if err := storage.DB.Table(tname).AutoMigrate(dts...); err != nil {
			panic(err)
		}
		th.Name = fmt.Sprintf("task_host_%d", id)
		dts = append(dts, th)
	}
}

type TaskHost struct {
	Ii     int64 `gorm:"primaryKey"`
	Id     int64
	Host   string `gorm:"column:severities;type:varchar(32);not null;default:''"`
	Status string `gorm:"column:status;type:varchar(32);not null;default:''"`
	Stdout string `gorm:"column:stdout;type:varchar(2048);not null;default:''"`
	Stderr string `gorm:"column:stderr;type:varchar(2048);not null;default:''"`
	Name   string `gorm:"-"`
}

func (t TaskHost) TableName() string {
	return t.Name
}
