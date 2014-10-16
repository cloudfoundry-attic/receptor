package api

import (
	"bytes"
	"encoding/json"

	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

func NewErrorResponse(err error) ErrorResponse {
	return ErrorResponse{err.Error()}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (res ErrorResponse) JSONReader() *bytes.Buffer {
	return writeJSON(res)
}

type CreateTaskRequest struct {
	TaskGuid   string                  `json:"task_guid"`
	Domain     string                  `json:"domain"`
	Actions    []models.ExecutorAction `json:"actions"`
	Stack      string                  `json:"stack"`
	MemoryMB   int                     `json:"memory_mb"`
	DiskMB     int                     `json:"disk_mb"`
	CpuPercent float64                 `json:"cpu_percent"`
	Log        models.LogConfig        `json:"log"`
	Annotation string                  `json:"annotation,omitempty"`
}

func (req CreateTaskRequest) ToTask() (models.Task, error) {
	task := models.Task{
		TaskGuid:   req.TaskGuid,
		Domain:     req.Domain,
		Actions:    req.Actions,
		Stack:      req.Stack,
		MemoryMB:   req.MemoryMB,
		DiskMB:     req.DiskMB,
		CpuPercent: req.CpuPercent,
		Log:        req.Log,
		Annotation: req.Annotation,
	}

	err := task.Validate()
	if err != nil {
		return models.Task{}, err
	}
	return task, nil
}

func (req CreateTaskRequest) JSONReader() *bytes.Buffer {
	return writeJSON(req)
}

type JSONReader interface {
	JSONReader() *bytes.Buffer
}

func writeJSON(jsonObj interface{}) *bytes.Buffer {
	jsonBytes, err := json.Marshal(jsonObj)
	if err != nil {
		panic("Unable to encode JSON: " + err.Error())
	}
	return bytes.NewBuffer(jsonBytes)
}
