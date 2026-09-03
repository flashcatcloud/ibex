package timer

import (
	"context"
	"fmt"
	"time"

	"github.com/flashcatcloud/ibex/src/models"

	"github.com/toolkits/pkg/logger"
)

// CacheHostDoing 缓存task_host_doing表全部内容，减轻DB压力
func CacheHostDoing() {
	if err := cacheHostDoing(); err != nil {
		fmt.Println("cannot cache task_host_doing data: ", err)
	}
	go loopCacheHostDoing()
}

func loopCacheHostDoing() {
	for {
		time.Sleep(time.Millisecond * 400)
		if err := cacheHostDoing(); err != nil {
			logger.Warning("cannot cache task_host_doing data: ", err)
		}
	}
}

func cacheHostDoing() error {
	// 任何一路取数失败都直接返回，保留上一轮的缓存。
	// 用残缺的数据覆盖缓存会让 agent 静默地收不到任务：Report 照常返回，只是 AssignTasks
	// 空了，排查时唯一的线索只有这里的一行日志。宁可短暂用旧数据，也不要下发一个空集合。
	doingsFromDb, err := models.TableRecordGets[[]models.TaskHostDoing](models.TaskHostDoing{}.TableName(), "")
	if err != nil {
		logger.Errorf("models.TableRecordGets fail: %v", err)
		return err
	}

	ctx := context.Background()

	doingsFromRedis, err := models.CacheRecordGets[models.TaskHostDoing](ctx)
	if err != nil {
		logger.Errorf("models.CacheRecordGets fail: %v", err)
		return err
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

	models.SetDoingCache(set)

	return err
}
