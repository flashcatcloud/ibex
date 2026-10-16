package models

import (
	"strings"

	"github.com/flashcatcloud/ibex/src/server/config"
	"github.com/flashcatcloud/ibex/src/types"
	"github.com/google/uuid"
)

// ProfilingResultStored verifies persistence rather than interpreting a nil
// MarkDoneStatus error as acknowledgement (some branches intentionally no-op).
// Edge/cache-only results are not acknowledged: retain them on the Agent.
func ProfilingResultStored(host string, r types.ReportTask) (bool, error) {
	if !config.C.IsCenter || r.Id <= 0 || r.Clock <= 0 || (r.Status != "success" && r.Status != "failed") {
		return false, nil
	}
	id, err := uuid.Parse(strings.TrimPrefix(r.ProfilingTaskID, "cpt-"))
	if err != nil || r.ProfilingTaskID != "cpt-"+id.String() {
		return false, nil
	}
	// Query the database, not the metadata/assignment caches.
	meta, err := TaskMetaGet("id = ?", r.Id)
	if err != nil || meta == nil {
		return false, err
	}
	if !profilingReceiptMatchesMeta(meta, r.ProfilingTaskID) {
		return false, nil
	}
	n, err := TableRecordCount(TaskHostDoing{}.TableName(), "id = ? and host = ?", r.Id, host)
	if err != nil || n != 0 {
		return false, err
	}
	stored, err := TaskHostGet(r.Id, host)
	if err != nil {
		return false, err
	}
	return profilingReceiptMatchesResult(stored, r), nil
}

func profilingReceiptMatchesMeta(meta *TaskMeta, parent string) bool {
	if meta == nil || meta.SystemCaller != "continuous-profiling" {
		return false
	}
	script := meta.Script
	if strings.HasPrefix(script, "#!") {
		p := strings.IndexByte(script, '\n')
		if p < 0 {
			return false
		}
		script = script[p+1:]
	}
	return strings.HasPrefix(script, "# CP_RETENTION_V1 task="+parent+"\n")
}

func profilingReceiptMatchesResult(stored *TaskHost, r types.ReportTask) bool {
	return stored != nil && stored.Status == r.Status && stored.Stdout == r.Stdout && stored.Stderr == r.Stderr
}
