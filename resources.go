package receptor

import "github.com/cloudfoundry-incubator/runtime-schema/models"

type LogConfig struct {
	Guid       string `json:"guid"`
	SourceName string `json:"source_name"`
}

type EnvironmentVariable struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type PortMapping struct {
	ContainerPort uint32 `json:"container_port"`
	HostPort      uint32 `json:"host_port,omitempty"`
}

const (
	TaskStateInvalid   = "INVALID"
	TaskStatePending   = "PENDING"
	TaskStateClaimed   = "CLAIMED"
	TaskStateRunning   = "RUNNING"
	TaskStateCompleted = "COMPLETED"
	TaskStateResolving = "RESOLVING"
)

type TaskCreateRequest struct {
	Actions               []models.ExecutorAction `json:"actions"`
	Annotation            string                  `json:"annotation,omitempty"`
	CompletionCallbackURL string                  `json:"completion_callback_url"`
	CPUWeight             uint                    `json:"cpu_weight"`
	DiskMB                int                     `json:"disk_mb"`
	Domain                string                  `json:"domain"`
	Log                   LogConfig               `json:"log"`
	MemoryMB              int                     `json:"memory_mb"`
	ResultFile            string                  `json:"result_file"`
	Stack                 string                  `json:"stack"`
	TaskGuid              string                  `json:"task_guid"`
	RootFSPath            string                  `json:"root_fs"`
	EnvironmentVariables  []EnvironmentVariable   `json:"env,omitempty"`
}

type TaskResponse struct {
	Actions               []models.ExecutorAction `json:"actions"`
	Annotation            string                  `json:"annotation,omitempty"`
	CompletionCallbackURL string                  `json:"completion_callback_url"`
	CPUWeight             uint                    `json:"cpu_weight"`
	DiskMB                int                     `json:"disk_mb"`
	Domain                string                  `json:"domain"`
	Log                   LogConfig               `json:"log"`
	MemoryMB              int                     `json:"memory_mb"`
	ResultFile            string                  `json:"result_file"`
	Stack                 string                  `json:"stack"`
	TaskGuid              string                  `json:"task_guid"`
	RootFSPath            string                  `json:"root_fs"`
	EnvironmentVariables  []EnvironmentVariable   `json:"env,omitempty"`

	CreatedAt     int64  `json:"created_at"`
	Failed        bool   `json:"failed"`
	FailureReason string `json:"failure_reason"`
	Result        string `json:"result"`
	State         string `json:"state"`
}

type DesiredLRPCreateRequest struct {
	ProcessGuid          string                  `json:"process_guid"`
	Domain               string                  `json:"domain"`
	RootFSPath           string                  `json:"root_fs"`
	Instances            int                     `json:"instances"`
	Stack                string                  `json:"stack"`
	EnvironmentVariables []EnvironmentVariable   `json:"env,omitempty"`
	Actions              []models.ExecutorAction `json:"actions"`
	DiskMB               int                     `json:"disk_mb"`
	MemoryMB             int                     `json:"memory_mb"`
	CPUWeight            uint                    `json:"cpu_weight"`
	Ports                []PortMapping           `json:"ports"`
	Routes               []string                `json:"routes"`
	Log                  LogConfig               `json:"log"`
	Annotation           string                  `json:"annotation,omitempty"`
}

type DesiredLRPUpdateRequest struct {
	Instances  *int     `json:"instances,omitempty"`
	Routes     []string `json:"routes,omitempty"`
	Annotation *string  `json:"annotation,omitempty"`
}

type DesiredLRPResponse struct {
	ProcessGuid          string                  `json:"process_guid"`
	Domain               string                  `json:"domain"`
	RootFSPath           string                  `json:"root_fs"`
	Instances            int                     `json:"instances"`
	Stack                string                  `json:"stack"`
	EnvironmentVariables []EnvironmentVariable   `json:"env,omitempty"`
	Actions              []models.ExecutorAction `json:"actions"`
	DiskMB               int                     `json:"disk_mb"`
	MemoryMB             int                     `json:"memory_mb"`
	CPUWeight            uint                    `json:"cpu_weight"`
	Ports                []PortMapping           `json:"ports"`
	Routes               []string                `json:"routes"`
	Log                  LogConfig               `json:"log"`
	Annotation           string                  `json:"annotation,omitempty"`
}
