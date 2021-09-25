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

func New(version string) *gin.Engine {
	gin.SetMode(config.C.RunMode)

	loggerMid := aop.Logger()
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

	api := r.Group("/ibex/v1", gin.BasicAuth(config.C.BasicAuth))
	{
		api.GET("/task/:id/stdout", taskStdout)
		api.GET("/task/:id/stderr", taskStderr)
		api.GET("/task/:id/state", taskState)
		api.GET("/task/:id/result", taskResult)
		api.GET("/task/:id/host/:host/output", taskHostOutput)
		api.GET("/task/:id/host/:host/stdout", taskHostStdout)
		api.GET("/task/:id/host/:host/stderr", taskHostStderr)
		api.GET("/task/:id/stdout.txt", taskStdoutTxt)
		api.GET("/task/:id/stderr.txt", taskStderrTxt)
		api.GET("/task/:id/stdout.json", taskStdoutJSON)
		api.GET("/task/:id/stderr.json", taskStderrJSON)
	}
}
