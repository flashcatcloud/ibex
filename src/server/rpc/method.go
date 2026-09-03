package rpc

import (
	"fmt"
	"os"
	"sync"

	"github.com/toolkits/pkg/logger"

	"github.com/flashcatcloud/ibex/src/models"
	"github.com/flashcatcloud/ibex/src/types"
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
	resp.SystemCaller = meta.SystemCaller

	return nil
}

// reporting 保证同一个 ident 同时只有一个回写 goroutine 在跑。
// agent 在拿到成功响应前会反复重发同一批结果，没有这层保护会堆积 goroutine。
var reporting sync.Map // ident -> struct{}

func (*Server) Report(req types.ReportRequest, resp *types.ReportResponse) error {
	// 结果回写必须与任务下发解耦。回写是逐条落库的，在 edge 上每条还要发一次 HTTP 到
	// center，跨机房部署时总耗时远超 agent 的 RPC 超时（5s）。如果同步做，一批回写不掉的
	// 结果会让 Report 迟迟不返回：agent 每次超时重连，本地结果因此永远清不掉，下一轮又原样
	// 重发，而 AssignTasks 一直发不出去——这台机器的任务通道就被永久堵死了。
	if len(req.ReportTasks) > 0 {
		if _, busy := reporting.LoadOrStore(req.Ident, struct{}{}); busy {
			// 上一批还在回写。这一批本次不落库，agent 侧 Clean 之后就取不回来了，
			// 所以这里必须留痕。正常情况下回写是毫秒级的，只有积压补写时才会走到。
			logger.Warningf("skip report of %d task(s) from %s: previous writeback still in flight",
				len(req.ReportTasks), req.Ident)
		} else {
			go func(r types.ReportRequest) {
				defer reporting.Delete(r.Ident)
				handleDoneTask(r)
			}(req)
		}
	}

	doings := models.GetDoingCache(req.Ident)

	tasks := make([]types.AssignTask, 0, len(doings))
	for _, doing := range doings {
		tasks = append(tasks, types.AssignTask{
			Id:     doing.Id,
			Clock:  doing.Clock,
			Action: doing.Action,
		})
	}
	resp.AssignTasks = tasks

	return nil
}

// handleDoneTask 逐条回写 agent 上报的执行结果。单条失败只记录日志并继续处理后面的，
// 不能中断整批：一条写不进去的结果会让它后面的所有结果都没有机会落库。
func handleDoneTask(req types.ReportRequest) {
	count := len(req.ReportTasks)
	val, ok := os.LookupEnv("CONTINUOUS_OUTPUT")
	for i := 0; i < count; i++ {
		t := req.ReportTasks[i]
		if ok && val == "1" && t.Status == "running" {
			err := models.RealTimeUpdateOutput(t.Id, req.Ident, t.Stdout, t.Stderr)
			if err != nil {
				logger.Errorf("cannot update output, id:%d, hostname:%s, clock:%d, status:%s, err: %v", t.Id, req.Ident, t.Clock, t.Status, err)
				continue
			}
		} else {
			if t.Status == "success" || t.Status == "failed" {
				_, isEdgeAlertTriggered := models.CheckExistAndEdgeAlertTriggered(req.Ident, t.Id)
				// ibex agent可能会重复上报结果，如果任务已经不在task_host_doing缓存中了，说明该任务已经MarkDone了，不需要再处理
				// if !exist {
				// 	continue
				// }
				// update:
				// 之前的代码的问题：如果脚本超时了，服务端会结束任务，客户端最终脚本执行的 stdout 和 stderr 虽然最终也上报了但是被服务端 continue 丢弃了
				// 改造之后，可以接收超时脚本的 stdout 和 stderr。
				// 但是：如果 edge 和中心之间网络断了，就会写失败

				err := models.MarkDoneStatus(t.Id, t.Clock, req.Ident, t.Status, t.Stdout, t.Stderr, isEdgeAlertTriggered)
				if err != nil {
					logger.Errorf("cannot mark task done, id:%d, hostname:%s, clock:%d, status:%s, err: %v", t.Id, req.Ident, t.Clock, t.Status, err)
					continue
				}
			}
		}

	}
}
