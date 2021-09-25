package router

import (
	"net/http"

	"github.com/toolkits/pkg/errorx"
	"github.com/ulricqin/ibex/src/models"
)

func TaskMeta(id int64) *models.TaskMeta {
	obj, err := models.TaskMetaGet("id = ?", id)
	errorx.Dangerous(err)

	if obj == nil {
		errorx.Bomb(http.StatusNotFound, "no such task meta")
	}

	return obj
}
