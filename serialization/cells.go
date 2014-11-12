package serialization

import (
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

func ExecutorPresenceToCellResponse(executorPresence models.ExecutorPresence) receptor.CellResponse {
	return receptor.CellResponse{
		CellID: executorPresence.ExecutorID,
		Stack:  executorPresence.Stack,
	}
}
