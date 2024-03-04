package router

import (
	"fmt"

	"os"
	"strings"

	"github.com/ulricqin/ibex/src/pkg/aop"
	"github.com/ulricqin/ibex/src/server/config"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
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

	configBaseRouter(r, version)
	ConfigRouter(r)

	return r
}

func configBaseRouter(r *gin.Engine, version string) {
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
}

func ConfigRouter(r *gin.Engine) {
	api := r.Group("/ibex/v1", gin.BasicAuth(config.C.BasicAuth))
	{
		api.POST("/tasks", taskAdd)
		api.GET("/tasks", taskGets)
		api.GET("/tasks/done-ids", doneIds)
		api.GET("/task/:id", taskGet)
		api.PUT("/task/:id/action", taskAction)
		api.GET("/task/:id/stdout", taskStdout)
		api.GET("/task/:id/stderr", taskStderr)
		api.GET("/task/:id/state", taskState)
		api.GET("/task/:id/result", taskResult)
		api.PUT("/task/:id/host/:host/action", taskHostAction)
		api.GET("/task/:id/host/:host/output", taskHostOutput)
		api.GET("/task/:id/host/:host/stdout", taskHostStdout)
		api.GET("/task/:id/host/:host/stderr", taskHostStderr)
		api.GET("/task/:id/stdout.txt", taskStdoutTxt)
		api.GET("/task/:id/stderr.txt", taskStderrTxt)
		api.GET("/task/:id/stdout.json", taskStdoutJSON)
		api.GET("/task/:id/stderr.json", taskStderrJSON)

		// api for edge server
		api.POST("/db/record/list", tableRecordListGet)
		api.POST("/db/record/count", tableRecordCount)
		api.POST("/mark/done", markDone)
		api.POST("/task/meta", taskMetaAdd)
		api.POST("/task/host/", taskHostAdd)
		api.POST("/task/hosts/upsert", taskHostUpsert)
	}
}
