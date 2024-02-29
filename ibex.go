package ibex

import "C"
import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/ulricqin/ibex/src/server/config"
	"github.com/ulricqin/ibex/src/server/router"
	"github.com/ulricqin/ibex/src/server/rpc"
	"github.com/ulricqin/ibex/src/server/timer"
	"github.com/ulricqin/ibex/src/storage"
	"gorm.io/gorm"
	"os"
	"strings"
)

func EdgeServerStart(cache redis.Cmdable, rpcListen string, api config.CenterApi, r *gin.Engine) {
	config.C.IsCenter = false
	config.C.CenterApi = api
	config.C.BasicAuth = make(gin.Accounts)
	config.C.BasicAuth[api.BasicAuthUser] = api.BasicAuthPass

	router.ConfigRouter(r)

	storage.Cache = cache

	rpc.Start(rpcListen)

	timer.CacheHostDoing()
	timer.ReportResult()
}

func CenterServerStart(db *gorm.DB, cache redis.Cmdable, rpcListen string, api config.CenterApi, r *gin.Engine) {
	config.C.IsCenter = true
	config.C.BasicAuth = make(gin.Accounts)
	config.C.BasicAuth[api.BasicAuthUser] = api.BasicAuthPass
	config.C.Heartbeat.LocalAddr = schedulerAddrGet(rpcListen)

	router.ConfigRouter(r)

	storage.DB = db
	storage.Cache = cache

	rpc.Start(rpcListen)

	timer.CacheHostDoing()
	timer.ReportResult()
	go timer.Heartbeat()
	go timer.Schedule()
	go timer.CleanLong()
}

func schedulerAddrGet(rpcListen string) string {
	ip := fmt.Sprint(config.GetOutboundIP())
	if ip == "" {
		fmt.Println("heartbeat ip auto got is blank")
		os.Exit(1)
	}

	port := strings.Split(rpcListen, ":")[1]
	localAddr := ip + ":" + port
	return localAddr
}
