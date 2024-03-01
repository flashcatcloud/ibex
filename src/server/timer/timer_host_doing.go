package timer

import (
	"context"
	"time"

	"github.com/ulricqin/ibex/src/models"

	"github.com/toolkits/pkg/logger"
)

// CacheHostDoing 缓存task_host_doing表全部内容，减轻DB压力
func CacheHostDoing() {
	cacheHostDoing()
	go loopCacheHostDoing()
}

func loopCacheHostDoing() {
	for {
		time.Sleep(time.Millisecond * 400)
		cacheHostDoing()
	}
}

func cacheHostDoing() {
	doingsFromDb, err := models.DBRecordList[[]models.TaskHostDoing](models.TaskHostDoing{}.TableName(), "")
	if err != nil {
		logger.Errorf("models.DBRecordList fail: %v", err)
	}

	ctx := context.Background()
	keys, err := models.CacheKeyList(ctx, "host:doing:*")
	if err != nil {
		logger.Errorf("models.CacheKeyList fail: %v", err)
	}
	doingsFromRedis, err := models.CacheRecordList[models.TaskHostDoing](ctx, keys)
	if err != nil {
		logger.Errorf("models.CacheRecordList fail: %v", err)
	}

	set := make(map[string][]models.TaskHostDoing)
	for _, doing := range doingsFromDb {
		doing.AlertTriggered = false
		set[doing.Host] = append(set[doing.Host], doing)
	}
	for _, doing := range doingsFromRedis {
		doing.AlertTriggered = true
		set[doing.Host] = append(set[doing.Host], doing)
	}

	models.SetDoingLocalCache(set)
}
