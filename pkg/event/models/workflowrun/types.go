package workflowrun

// Data contains WorkflowJob breif information and WorkflowRun detail.
type Data struct {
	ParentFullName string      `json:"parentFullName"`
	ProjectName    string      `json:"projectName"`
	IsMultiBranch  bool        `json:"multiBranch"`
	Run            WorkflowRun `json:"run"`
}

// WorkflowRun contains WorkflowRun detail.
type WorkflowRun struct {
	Actions           Actions `json:"actions"`
	Building          bool    `json:"building"`
	Description       string  `json:"description"`
	DisplayName       string  `json:"displayName"`
	Duration          int     `json:"duration"`
	EstimatedDuration int     `json:"estimatedDuration"`
	FullDisplayName   string  `json:"fullDisplayName"`
	ID                string  `json:"id"`
	KeepLog           bool    `json:"keepLog"`
	Number            int     `json:"number"`
	QueueID           int     `json:"queueId"`
	Result            string  `json:"result"`
	Timestamp         int64   `json:"timestamp"`
}

// Funcs is a collection of handlers for various event type.
type Funcs struct {
	HandleInitialize func(*Data) error
	HandleStarted    func(*Data) error
	HandleFinalized  func(*Data) error
	HandleCompleted  func(*Data) error
	HandleDeleted    func(*Data) error
}
