package models

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

type TaskHostDoing struct {
	Id     int64 `gorm:"primaryKey"`
	Host   string
	Clock  int64
	Action string
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

func SetDoingLocalCache(v map[string][]TaskHostDoing) {
	doingLock.Lock()
	doingMaps = v
	doingLock.Unlock()
}

func GetDoingLocalCache(host string) []TaskHostDoing {
	doingLock.RLock()
	defer doingLock.RUnlock()

	return doingMaps[host]
}

func GetDoingRedisCache(host string) ([]TaskHostDoing, error) {
	ctx := context.Background()

	keys, err := CacheKeyList(ctx, fmt.Sprintf("host:doing:%s", host))
	if err != nil {
		return nil, err
	}

	lst, err := CacheRecordList[TaskHostDoing](ctx, keys)

	return lst, err
}
