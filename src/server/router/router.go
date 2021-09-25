package router

import (
	"fmt"
	"os"
	"strings"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"

	"github.com/ulricqin/ibex/src/pkg/aop"
	"github.com/ulricqin/ibex/src/server/config"
)

var InternalServerError = "InternalServerError"

var skipPaths = []string{}

func New(version string) *gin.Engine {
	gin.SetMode(config.C.RunMode)

	loggerMid := aop.LoggerWithConfig(aop.LoggerConfig{SkipPaths: skipPaths})
	recoveryMid := aop.Recovery()

	if strings.ToLower(config.C.RunMode) == "release" {
		aop.DisableConsoleColor()
	}

	r := gin.New()

	r.Use(recoveryMid)

	// whether print access log
	if config.C.HTTP.PrintAccessLog {
		r.Use(loggerMid)
	}

	configRoute(r, version)

	return r
}

func configRoute(r *gin.Engine, version string) {
	if config.C.HTTP.PProf {
		pprof.Register(r, "/debug/pprof")
	}

	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	r.GET("/pid", func(c *gin.Context) {
		c.String(200, fmt.Sprintf("%d", os.Getpid()))
	})

	r.GET("/addr", func(c *gin.Context) {
		c.String(200, c.Request.RemoteAddr)
	})

	r.GET("/version", func(c *gin.Context) {
		c.String(200, version)
	})

	api := r.Group("/v1", gin.BasicAuth(config.C.BasicAuth))
	{
		api.GET("/version", func(c *gin.Context) {
			c.String(200, version)
		})
	}
}
