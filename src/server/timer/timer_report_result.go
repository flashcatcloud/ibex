package timer

import (
	"context"
	"fmt"
	"github.com/toolkits/pkg/logger"
	"github.com/ulricqin/ibex/src/models"
	"time"
)

func ReportResult() {
	if err := models.ReportCacheResult(context.Background()); err != nil {
		fmt.Println("cannot report task_host result from alter trigger: ", err)
	}
	go loopReport()
}

func loopReport() {
	d := time.Duration(2) * time.Second
	for {
		time.Sleep(d)
		if err := models.ReportCacheResult(context.Background()); err != nil {
			logger.Warning("cannot report task_host result from alter trigger: ", err)
		}
	}
}
