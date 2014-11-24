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
	TaskNotDeletable      = "TaskNotDeletable"
	TaskNotFound          = "TaskNotFound"
	InvalidTask           = "InvalidTask"

	DesiredLRPNotFound = "DesiredLRPNotFound"
	InvalidLRP         = "InvalidLRP"

	InvalidFreshness = "InvalidFreshness"

	InvalidJSON    = "InvalidJSON"
	InvalidRequest = "InvalidRequest"

	UnknownError = "UnknownError"
	Unauthorized = "Unauthorized"

	ActualLRPIndexNotFound = "ActualLRPIndexNotFound"
)
