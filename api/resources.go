package api

import (
	"encoding/json"
	"io"

	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

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

func (req CreateTaskRequest) JSONReader() io.Reader {
	pipeReader, pipeWriter := io.Pipe()
	jsonEncoder := json.NewEncoder(pipeWriter)
	go func() {
		jsonEncoder.Encode(req)
		pipeWriter.Close()
	}()
	return pipeReader
}
