package models

import (
	"fmt"
	"github.com/ulricqin/ibex/src/pkg/poster"
	"github.com/ulricqin/ibex/src/server/config"
	"gorm.io/gorm"

	"github.com/ulricqin/ibex/src/storage"
)

func DB() *gorm.DB {
	return storage.DB
}

func Count(tx *gorm.DB) (int64, error) {
	var cnt int64
	err := tx.Count(&cnt).Error
	return cnt, err
}

func Exists(tx *gorm.DB) (bool, error) {
	num, err := Count(tx)
	return num > 0, err
}

func Insert(objPtr interface{}) error {
	return DB().Create(objPtr).Error
}

func tht(id int64) string {
	return fmt.Sprintf("task_host_%d", id%100)
}

func DBRecordList[T any](table, where string, args ...interface{}) (T, error) {
	var lst T
	if config.C.IsCenter {
		err := DB().Table(table).Where(where, args...).Find(&lst).Error
		return lst, err
	}

	return poster.PostByUrlsWithResp[T](config.C.CenterApi, "/ibex/v1/record/list", map[string]interface{}{
		"table": table,
		"where": where,
		"args":  args,
	})
}

func DBRecordCount(table, where string, args ...interface{}) (int64, error) {
	if config.C.IsCenter {
		return Count(DB().Table(table).Where(where, args...))
	}

	return poster.PostByUrlsWithResp[int64](config.C.CenterApi, "/ibex/v1/record/count", map[string]interface{}{
		"table": table,
		"where": where,
		"args":  args,
	})
}

func DBRecordUpdate(values map[string]interface{}, table string, where string, args ...interface{}) error {
	if config.C.IsCenter {
		return DB().Table(table).Where(where, args...).Updates(values).Error
	}

	return poster.PostByUrls(config.C.CenterApi, "/ibex/v1/record/update", values)
}

func getSqlCountPath(table, where string, args ...interface{}) string {
	path := "/ibex/v1/sql/count"
	path += "?where=" + where
	path += "&args=" + fmt.Sprintf("%v", args)
	path += "table=" + table
	return path
}
