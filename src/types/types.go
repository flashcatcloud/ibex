package types

type TaskMetaResponse struct {
	Message      string
	Script       string
	Args         string
	Account      string
	Stdin        string
	SystemCaller string
}

type ReportTask struct {
	ProfilingTaskID string
	Id              int64
	Clock           int64
	Status          string
	Stdout          string
	Stderr          string
}

type ReportRequest struct {
	Ident       string
	ReportTasks []ReportTask
}

type AssignTask struct {
	Id     int64
	Clock  int64
	Action string
}

type ProfilingReceipt struct {
	Id              int64
	Clock           int64
	ProfilingTaskID string
}

type ReportResponse struct {
	ProfilingReceipts []ProfilingReceipt
	Message           string
	AssignTasks       []AssignTask
}
