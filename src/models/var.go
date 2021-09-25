package models

import (
	"gorm.io/gorm"
	
	"github.com/toolkits/pkg/str"
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
