package ibex

import (
	"fmt"
	"os"
	"strings"

	"github.com/ulricqin/ibex/src/server/config"
	"github.com/ulricqin/ibex/src/server/router"
	"github.com/ulricqin/ibex/src/server/rpc"
	"github.com/ulricqin/ibex/src/server/timer"
	"github.com/ulricqin/ibex/src/storage"

	n9eConf "github.com/ccfos/nightingale/v6/conf"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var (
	Conf     *n9eConf.Ibex
	HttpPort int
)

func ServerStart(isCenter bool, db *gorm.DB, rc redis.Cmdable, n9eIbex n9eConf.Ibex, api *n9eConf.CenterApi, r *gin.Engine, httpPort int) {
	Conf = &n9eIbex

	config.C.IsCenter = isCenter
	config.C.BasicAuth = make(gin.Accounts)
	config.C.BasicAuth[n9eIbex.BasicAuthUser] = n9eIbex.BasicAuthPass
	config.C.Heartbeat.LocalAddr = schedulerAddrGet(n9eIbex.RPCListen)
	HttpPort = httpPort

	router.ConfigRouter(r)

	storage.Cache = rc
	if err := storage.IdInit(); err != nil {
		fmt.Println("cannot init id generator: ", err)
		os.Exit(1)
	}

	rpc.Start(n9eIbex.RPCListen)

	timer.CacheHostDoing()
	timer.ReportResult()

	if isCenter {
		storage.DB = db
		go timer.Heartbeat()
		go timer.Schedule()
		go timer.CleanLong()
	} else {
		config.C.CenterApi = *api
	}

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
