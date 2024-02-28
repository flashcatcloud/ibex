package ibex

import (
	"github.com/redis/go-redis/v9"
	"github.com/ulricqin/ibex/src/server/config"
	"github.com/ulricqin/ibex/src/server/rpc"
	"github.com/ulricqin/ibex/src/server/timer"
	"github.com/ulricqin/ibex/src/storage"
	"gorm.io/gorm"
)

func EdgeServerStart(db *gorm.DB, cache redis.Cmdable, rpcListen string, api config.CenterApi) {
	config.C.IsCenter = false
	config.C.CenterApi = api

	storage.DB = db
	storage.Cache = cache

	rpc.Start(rpcListen)

	timer.CacheHostDoing()
	timer.ReportResult()
}

func CenterServerStart(db *gorm.DB, cache redis.Cmdable, rpcListen string) {
	config.C.IsCenter = true

	storage.DB = db
	storage.Cache = cache

	rpc.Start(rpcListen)

	timer.CacheHostDoing()
	timer.ReportResult()
	go timer.Heartbeat()
	go timer.Schedule()
	go timer.CleanLong()
}