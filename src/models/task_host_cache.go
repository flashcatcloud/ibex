package models

import (
	"github.com/toolkits/pkg/logger"
	"github.com/ulricqin/ibex/src/storage"
	"sync"
)

type TaskHostCacheType struct {
	cache []TaskHost
	sync.RWMutex
}

const initialSize = 128

var taskHostCache *TaskHostCacheType

func InitTaskHostCache() {
	taskHostCache = new(TaskHostCacheType)
	taskHostCache.cache = make([]TaskHost, 0, initialSize)
}

func (thc *TaskHostCacheType) Set(th TaskHost) {
	thc.Lock()
	defer thc.Unlock()

	thc.cache = append(thc.cache, th)
}

func (t *TaskHostCacheType) PopAll() []TaskHost {
	t.Lock()
	defer t.Unlock()

	all := t.cache
	t.cache = make([]TaskHost, 0, initialSize)

	return all
}

func ReportCacheResult() error {
	result := taskHostCache.PopAll()
	dones := make([]TaskHost, 0)
	for _, th := range result {
		// id大于redis初始id，说明是edge与center失联时，本地告警规则触发的自愈脚本，生成的id
		// 为了防止不同边缘机房生成的脚本任务id相同，不上报结果至数据库
		if th.Id >= storage.IDINITIAL {
			logger.Infof("task[%s] host[%s] done, result:[%v]", th.Id, th.Host, th)
		} else {
			dones = append(dones, th)
		}
	}

	if len(dones) == 0 {
		return nil
	}

	errs, err := TaskHostUpserts(dones)
	if err != nil {
		return err
	}
	for key, err := range errs {
		logger.Warningf("report task_host_cache[%s] result error: %v", key, err)
	}
	return nil
}

//func ReportCacheResult(ctx context.Context) error {
//	keys, err := CacheKeyGets(ctx, "task:host:*")
//	if err != nil {
//		return err
//	}
//
//	lst, err := CacheRecordGets[TaskHost](ctx, keys)
//	if err != nil {
//		return err
//	}
//
//	dones := make([]TaskHost, 0)
//	for _, task := range lst {
//		if task.Status != "running" {
//			dones = append(dones, task)
//		}
//	}
//	if len(dones) == 0 {
//		return nil
//	}
//
//	errs, err := TaskHostUpserts(dones)
//	if err != nil {
//		return err
//	}
//	for key, err := range errs {
//		logger.Warningf("report cache[%s] result error: %s", key, err.Error())
//	}
//	return nil
//}
