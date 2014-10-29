package receptor

import "github.com/cloudfoundry-incubator/runtime-schema/models"

const (
	TaskStateInvalid   = "INVALID"
	TaskStatePending   = "PENDING"
	TaskStateClaimed   = "CLAIMED"
	TaskStateRunning   = "RUNNING"
	TaskStateCompleted = "COMPLETED"
	TaskStateResolving = "RESOLVING"
)

type CreateTaskRequest struct {
	Actions               []models.ExecutorAction `json:"actions"`
	Annotation            string                  `json:"annotation,omitempty"`
	CompletionCallbackURL string                  `json:"completion_callback_url"`
	CPUWeight             uint                    `json:"cpu_weight"`
	DiskMB                int                     `json:"disk_mb"`
	Domain                string                  `json:"domain"`
	Log                   models.LogConfig        `json:"log"`
	MemoryMB              int                     `json:"memory_mb"`
	ResultFile            string                  `json:"result_file"`
	Stack                 string                  `json:"stack"`
	TaskGuid              string                  `json:"task_guid"`
	RootFSPath            string                  `json:"root_fs"`
}

type TaskResponse struct {
	Actions               []models.ExecutorAction `json:"actions"`
	Annotation            string                  `json:"annotation,omitempty"`
	CompletionCallbackURL string                  `json:"completion_callback_url"`
	CPUWeight             uint                    `json:"cpu_weight"`
	DiskMB                int                     `json:"disk_mb"`
	Domain                string                  `json:"domain"`
	Log                   models.LogConfig        `json:"log"`
	MemoryMB              int                     `json:"memory_mb"`
	ResultFile            string                  `json:"result_file"`
	Stack                 string                  `json:"stack"`
	TaskGuid              string                  `json:"task_guid"`
	RootFSPath            string                  `json:"root_fs"`

	CreatedAt     int64  `json:"created_at"`
	Failed        bool   `json:"failed"`
	FailureReason string `json:"failure_reason"`
	Result        string `json:"result"`
	State         string `json:"state"`
}
