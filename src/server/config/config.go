package config

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/koding/multiconfig"

	"github.com/ulricqin/ibex/src/pkg/httpx"
	"github.com/ulricqin/ibex/src/pkg/logx"
	"github.com/ulricqin/ibex/src/storage"
)

var (
	C    = new(Config)
	once sync.Once
)

func MustLoad(fpaths ...string) {
	once.Do(func() {
		loaders := []multiconfig.Loader{
			&multiconfig.TagLoader{},
			&multiconfig.EnvironmentLoader{},
		}

		for _, fpath := range fpaths {
			handled := false

			if strings.HasSuffix(fpath, "toml") {
				loaders = append(loaders, &multiconfig.TOMLLoader{Path: fpath})
				handled = true
			}
			if strings.HasSuffix(fpath, "conf") {
				loaders = append(loaders, &multiconfig.TOMLLoader{Path: fpath})
				handled = true
			}
			if strings.HasSuffix(fpath, "json") {
				loaders = append(loaders, &multiconfig.JSONLoader{Path: fpath})
				handled = true
			}
			if strings.HasSuffix(fpath, "yaml") {
				loaders = append(loaders, &multiconfig.YAMLLoader{Path: fpath})
				handled = true
			}

			if !handled {
				fmt.Println("config file invalid, valid file exts: .conf,.yaml,.toml,.json")
				os.Exit(1)
			}
		}

		m := multiconfig.DefaultLoader{
			Loader:    multiconfig.MultiLoader(loaders...),
			Validator: multiconfig.MultiValidator(&multiconfig.RequiredValidator{}),
		}
		m.MustLoad(C)
	})
}

type Config struct {
	RunMode   string
	RPC       RPC
	Heartbeat Heartbeat
	Output    Output
	Log       logx.Config
	HTTP      httpx.Config
	BasicAuth gin.Accounts
	Gorm      storage.Gorm
	MySQL     storage.MySQL
	Postgres  storage.Postgres
}

type RPC struct {
	Listen string
}

type Heartbeat struct {
	IP        string
	LocalAddr string
	Interval  int64
}

type Output struct {
	ComeFrom string
	AgtdPort int
}

func (c *Config) IsDebugMode() bool {
	return c.RunMode == "debug"
}
