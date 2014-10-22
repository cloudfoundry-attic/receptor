package receptor

type Error struct {
	Type    string `json:"name"`
	Message string `json:"message"`
}

func (err Error) Error() string {
	return err.Message
}

const (
	TaskGuidAlreadyExists = "TaskGuidAlreadyExists"
	TaskNotFound          = "TaskNotFound"
	InvalidJSON           = "InvalidJSON"
	InvalidTask           = "InvalidTask"
	UnknownError          = "UnknownError"
	Unauthorized          = "Unauthorized"
)
