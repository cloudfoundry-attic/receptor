package serialization

import (
	"net/url"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

func TaskFromRequest(req receptor.TaskCreateRequest) (models.Task, error) {
	var u *url.URL
	if req.CompletionCallbackURL != "" {
		var err error
		u, err = url.ParseRequestURI(req.CompletionCallbackURL)
		if err != nil {
			return models.Task{}, err
		}
	}

	task := models.Task{
		TaskGuid:              req.TaskGuid,
		Domain:                req.Domain,
		RootFSPath:            req.RootFSPath,
		Actions:               req.Actions,
		Privileged:            req.Privileged,
		Stack:                 req.Stack,
		MemoryMB:              req.MemoryMB,
		DiskMB:                req.DiskMB,
		CPUWeight:             req.CPUWeight,
		Log:                   LogConfigToModel(req.Log),
		ResultFile:            req.ResultFile,
		Annotation:            req.Annotation,
		CompletionCallbackURL: u,
		EnvironmentVariables:  EnvironmentVariablesToModel(req.EnvironmentVariables),
	}

	err := task.Validate()
	if err != nil {
		return models.Task{}, err
	}
	return task, nil
}

func TaskToResponse(task models.Task) receptor.TaskResponse {
	url := ""
	if task.CompletionCallbackURL != nil {
		url = task.CompletionCallbackURL.String()
	}

	return receptor.TaskResponse{
		TaskGuid:              task.TaskGuid,
		Domain:                task.Domain,
		RootFSPath:            task.RootFSPath,
		Actions:               task.Actions,
		Privileged:            task.Privileged,
		Stack:                 task.Stack,
		MemoryMB:              task.MemoryMB,
		DiskMB:                task.DiskMB,
		CPUWeight:             task.CPUWeight,
		Log:                   LogConfigFromModel(task.Log),
		Annotation:            task.Annotation,
		CompletionCallbackURL: url,
		EnvironmentVariables:  EnvironmentVariablesFromModel(task.EnvironmentVariables),

		CreatedAt:     task.CreatedAt,
		FailureReason: task.FailureReason,
		Failed:        task.Failed,
		Result:        task.Result,
		State:         taskStateToResponseState(task.State),
	}
}

func taskStateToResponseState(state models.TaskState) string {
	switch state {
	case models.TaskStateInvalid:
		return receptor.TaskStateInvalid
	case models.TaskStatePending:
		return receptor.TaskStatePending
	case models.TaskStateClaimed:
		return receptor.TaskStateClaimed
	case models.TaskStateRunning:
		return receptor.TaskStateRunning
	case models.TaskStateCompleted:
		return receptor.TaskStateCompleted
	case models.TaskStateResolving:
		return receptor.TaskStateResolving
	}

	return ""
}
