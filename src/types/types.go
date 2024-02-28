package types

type TaskMetaResponse struct {
	Message string
	Script  string
	Args    string
	Account string
	Stdin   string
}

type ReportTask struct {
	Id             int64
	Clock          int64
	Status         string
	Stdout         string
	Stderr         string
	AlertTriggered bool
}

type ReportRequest struct {
	Ident       string
	ReportTasks []ReportTask
}

type AssignTask struct {
	Id             int64
	Clock          int64
	Action         string
	AlertTriggered bool
}

type ReportResponse struct {
	Message     string
	AssignTasks []AssignTask
}
