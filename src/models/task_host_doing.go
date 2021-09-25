package models

import "sync"

type TaskHostDoing struct {
	Id     int64 `gorm:"primaryKey"`
	Host   string
	Clock  int64
	Action string
}

func (TaskHostDoing) TableName() string {
	return "task_host_doing"
}

func DoingHostList(where string, args ...interface{}) ([]TaskHostDoing, error) {
	var objs []TaskHostDoing
	err := DB().Where(where, args...).Find(&objs).Error
	return objs, err
}

func DoingHostCount(where string, args ...interface{}) (int64, error) {
	return Count(DB().Model(&TaskHostDoing{}).Where(where, args...))
}

var (
	doingLock sync.RWMutex
	doingMaps map[string][]TaskHostDoing
)

func SetDoingCache(v map[string][]TaskHostDoing) {
	doingLock.Lock()
	doingMaps = v
	doingLock.Unlock()
}

func GetDoingCache(k string) []TaskHostDoing {
	doingLock.RLock()
	defer doingLock.RUnlock()
	return doingMaps[k]
}
