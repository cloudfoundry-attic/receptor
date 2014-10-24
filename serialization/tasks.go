package serialization

import (
	"net/url"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

func TaskFromRequest(req receptor.CreateTaskRequest) (models.Task, error) {
	var url *url.URL
	if req.CompletionCallbackURL != "" {
		var err error
		url, err = url.Parse(req.CompletionCallbackURL)
		if err != nil {
			return models.Task{}, err
		}
	}

	task := models.Task{
		TaskGuid:              req.TaskGuid,
		Domain:                req.Domain,
		Actions:               req.Actions,
		Stack:                 req.Stack,
		MemoryMB:              req.MemoryMB,
		DiskMB:                req.DiskMB,
		CpuPercent:            req.CpuPercent,
		Log:                   req.Log,
		ResultFile:            req.ResultFile,
		Annotation:            req.Annotation,
		CompletionCallbackURL: url,
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
		Actions:               task.Actions,
		Stack:                 task.Stack,
		MemoryMB:              task.MemoryMB,
		DiskMB:                task.DiskMB,
		CpuPercent:            task.CpuPercent,
		Log:                   task.Log,
		Annotation:            task.Annotation,
		CompletionCallbackURL: url,
	}
}
