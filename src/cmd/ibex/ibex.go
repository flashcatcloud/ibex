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

	"github.com/ccfos/nightingale/v6/alert/aconf"
	n9eRouter "github.com/ccfos/nightingale/v6/center/router"
	"github.com/ccfos/nightingale/v6/conf"
	n9eConf "github.com/ccfos/nightingale/v6/conf"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var (
	HttpPort int
)

func ServerStart(isCenter bool, db *gorm.DB, rc redis.Cmdable, basicAuth gin.Accounts, heartbeat aconf.HeartbeatConfig,
	api *n9eConf.CenterApi, r *gin.Engine, centerRouter *n9eRouter.Router, ibex conf.Ibex, httpPort int) {
	config.C.IsCenter = isCenter
	config.C.BasicAuth = make(gin.Accounts)
	if len(basicAuth) > 0 {
		config.C.BasicAuth = basicAuth
	}

	config.C.Heartbeat.IP = heartbeat.IP
	config.C.Heartbeat.Interval = heartbeat.Interval
	config.C.Heartbeat.LocalAddr = schedulerAddrGet(ibex.RPCListen)
	HttpPort = httpPort

	config.C.Output.ComeFrom = ibex.Output.ComeFrom
	config.C.Output.AgtdPort = ibex.Output.AgtdPort

	if centerRouter != nil {
		router.ConfigRouter(r, centerRouter)
	} else {
		router.ConfigRouter(r)
	}

	storage.Cache = rc
	if err := storage.IdInit(); err != nil {
		fmt.Println("cannot init id generator: ", err)
		os.Exit(1)
	}
	if isCenter {
		storage.DB = db
	}

	if !isCenter {
		config.C.CenterApi = *api
	}

	rpc.Start(ibex.RPCListen)

	timer.CacheHostDoing()
	timer.ReportResult()
	if isCenter {
		go timer.Heartbeat()
		go timer.Schedule()
		go timer.CleanLong()
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
