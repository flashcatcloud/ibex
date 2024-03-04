package models

import (
	"encoding/json"
	"fmt"
	"sync"
)

type TaskHostDoing struct {
	Id             int64 `gorm:"primaryKey"`
	Host           string
	Clock          int64
	Action         string
	AlertTriggered bool `gorm:"-"`
}

func (TaskHostDoing) TableName() string {
	return "task_host_doing"
}

func (t *TaskHostDoing) MarshalBinary() ([]byte, error) {
	return json.Marshal(t)
}

func (t *TaskHostDoing) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, t)
}

func hostDoingCacheKey(id int64, host string) string {
	return fmt.Sprintf("host:doing:%s:%d", host, id)
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

func GetDoingCache(host string) []TaskHostDoing {
	doingLock.RLock()
	defer doingLock.RUnlock()

	return doingMaps[host]
}

func IsAlertTriggered(host string, id int64) (exist, isAlertTriggered bool) {
	doingLock.RLock()
	defer doingLock.RUnlock()

	hostDoing := doingMaps[host]
	for _, doing := range hostDoing {
		if doing.Id == id {
			exist = true
			isAlertTriggered = doing.AlertTriggered
			return
		}
	}

	return false, false
}
