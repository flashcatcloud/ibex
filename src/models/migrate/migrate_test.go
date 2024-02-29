package migrate

import (
	"fmt"
	"github.com/ccfos/nightingale/v6/pkg/ormx"
	"github.com/ulricqin/ibex/src/storage"
	"testing"
)

func TestMigrate(t *testing.T) {
	err := storage.InitDB(ormx.DBConfig{
		Debug:        true,
		DBType:       "mysql",
		DSN:          "root:1234@tcp(192.168.127.151:3306)/ibex?charset=utf8mb4&parseTime=True&loc=Local&allowNativePasswords=true",
		MaxLifetime:  7200,
		MaxIdleConns: 50,
		MaxOpenConns: 150,
		TablePrefix:  "",
	})
	if err != nil {
		fmt.Println(err)
	}
	Migrate()
}
