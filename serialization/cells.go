package serialization

import (
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

func CellPresenceToCellResponse(cellPresence models.CellPresence) receptor.CellResponse {
	return receptor.CellResponse{
		CellID: cellPresence.CellID,
		Stack:  cellPresence.Stack,
	}
}
