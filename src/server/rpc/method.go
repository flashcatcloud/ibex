package rpc

import (
	"fmt"

	"github.com/ulricqin/ibex/src/models"
	"github.com/ulricqin/ibex/src/types"

	"github.com/toolkits/pkg/logger"
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
	resp.Stdin = meta.Stdin

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

	hosts := models.GetDoingLocalCache(req.Ident)

	tasks := make([]types.AssignTask, 0, len(hosts))
	for _, h := range hosts {
		tasks = append(tasks, types.AssignTask{
			Id:     h.Id,
			Clock:  h.Clock,
			Action: h.Action,
		})
	}
	resp.AssignTasks = tasks

	return nil
}

func handleDoneTask(req types.ReportRequest) error {
	count := len(req.ReportTasks)
	for i := 0; i < count; i++ {
		t := req.ReportTasks[i]
		exist, isAlertTriggered := models.IsAlertTriggered(req.Ident, t.Id)
		// ibex agent可能会重复上报结果，如果任务已经不在task_host_doing缓存中了，说明该任务已经MarkDone了，不需要再处理
		if !exist {
			continue
		}

		err := models.MarkDoneStatus(t.Id, t.Clock, req.Ident, t.Status, t.Stdout, t.Stderr, isAlertTriggered)
		if err != nil {
			logger.Errorf("cannot mark task done, id:%d, hostname:%s, clock:%d, status:%s, err: %v", t.Id, req.Ident, t.Clock, t.Status, err)
			return err
		}
	}

	return nil
}
