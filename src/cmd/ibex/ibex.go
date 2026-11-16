package ibex

import (
	"fmt"
	"net"
	"os"

	"github.com/flashcatcloud/ibex/src/server/config"
	"github.com/flashcatcloud/ibex/src/server/router"
	"github.com/flashcatcloud/ibex/src/server/rpc"
	"github.com/flashcatcloud/ibex/src/server/timer"
	"github.com/flashcatcloud/ibex/src/storage"

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
	api *n9eConf.CenterApi, r *gin.Engine, centerRouter *n9eRouter.Router, ibex conf.Ibex, httpPort int, additional ...rpc.RegisterService) error {
	config.C.IsCenter = isCenter
	config.C.BasicAuth = make(gin.Accounts)
	if len(basicAuth) > 0 {
		config.C.BasicAuth = basicAuth
	}

	config.C.Heartbeat.IP = heartbeat.IP
	config.C.Heartbeat.Interval = heartbeat.Interval
	_, port, err := net.SplitHostPort(ibex.RPCListen)
	if err != nil {
		return fmt.Errorf("invalid ibex RPC listen address %q: %w", ibex.RPCListen, err)
	}
	config.C.Heartbeat.LocalAddr = net.JoinHostPort(config.GetOutboundIP().String(), port)
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
	} else {
		config.C.CenterApi = *api
	}

	if err := rpc.Start(ibex.RPCListen, additional...); err != nil {
		return err
	}
	if isCenter {
		go timer.Heartbeat()
		go timer.Schedule()
		go timer.CleanLong()
	}

	timer.CacheHostDoing()
	timer.ReportResult()
	return nil
}
