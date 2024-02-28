package models

import (
	"context"
	"fmt"
	"github.com/ulricqin/ibex/src/pkg/poster"
	"github.com/ulricqin/ibex/src/server/config"
	"github.com/ulricqin/ibex/src/storage"
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

func hostDoingCacheKey(id int64, host string) string {
	return fmt.Sprintf("host:doing:%s:%d", host, id)
}

func DoingHostList(where string, args ...interface{}) (lst []TaskHostDoing, err error) {
	if config.C.IsCenter {
		err = DB().Where(where, args...).Find(&lst).Error
	} else {
		path := getSqlCountPath(TaskHostDoing{}.TableName(), where, args...)
		lst, err = poster.GetByUrls[[]TaskHostDoing](config.C.CenterApi, path)
	}
	return
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
	iter := storage.Cache.Scan(ctx, 0, fmt.Sprintf("host:doing:%s", host), 0).Iterator()
	keys := make([]string, 0)
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}

	lst := make([]TaskHostDoing, 0, len(keys))
	err := storage.Cache.MGet(ctx, keys...).Scan(&lst)
	return nil, err
}
