package rpc

import (
	"fmt"

	"github.com/toolkits/pkg/logger"
	"github.com/ulricqin/ibex/src/models"
	"github.com/ulricqin/ibex/src/types"
)

// Ping return string 'pong', just for test
func (*Server) Ping(input string, output *string) error {
	*output = "pong"
	return nil
}

func (*Server) GetTaskMeta(id int64, resp *types.TaskMetaResponse) error {
	meta, err := models.TaskMetaGetByID(id)
	if err != nil {
		resp.Message = err.Error()
		return nil
	}

	if meta == nil {
		resp.Message = fmt.Sprintf("task %d not found", id)
		return nil
	}

	resp.Script = meta.Script
	resp.Args = meta.Args
	resp.Account = meta.Account
	resp.EventTags = meta.EventTags

	return nil
}

func (*Server) Report(req types.ReportRequest, resp *types.ReportResponse) error {
	if req.ReportTasks != nil && len(req.ReportTasks) > 0 {
		err := handleDoneTask(req)
		if err != nil {
			resp.Message = err.Error()
			return nil
		}
	}

	hosts := models.GetDoingCache(req.Ident)
	l := len(hosts)
	tasks := make([]types.AssignTask, l)
	for i := 0; i < l; i++ {
		tasks[i].Id = hosts[i].Id
		tasks[i].Clock = hosts[i].Clock
		tasks[i].Action = hosts[i].Action
	}

	resp.AssignTasks = tasks
	return nil
}

func handleDoneTask(req types.ReportRequest) error {
	count := len(req.ReportTasks)
	for i := 0; i < count; i++ {
		t := req.ReportTasks[i]
		err := models.MarkDoneStatus(t.Id, t.Clock, req.Ident, t.Status, t.Stdout, t.Stderr)
		if err != nil {
			logger.Errorf("cannot mark task done, id:%d, hostname:%s, clock:%d, status:%s", t.Id, req.Ident, t.Clock, t.Status)
			return err
		}
	}

	return nil
}
