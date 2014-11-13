package serialization

import (
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

func ActualLRPToResponse(actualLRP models.ActualLRP) receptor.ActualLRPResponse {
	return receptor.ActualLRPResponse{
		ProcessGuid:  actualLRP.ProcessGuid,
		InstanceGuid: actualLRP.InstanceGuid,
		CellID:       actualLRP.CellID,
		Domain:       actualLRP.Domain,
		Index:        actualLRP.Index,
		Host:         actualLRP.Host,
		Ports:        PortMappingFromModel(actualLRP.Ports),
		State:        actualLRPStateToResponseState(actualLRP.State),
		Since:        actualLRP.Since,
	}
}

func actualLRPStateToResponseState(state models.ActualLRPState) string {
	switch state {
	case models.ActualLRPStateInvalid:
		return receptor.ActualLRPStateInvalid
	case models.ActualLRPStateStarting:
		return receptor.ActualLRPStateStarting
	case models.ActualLRPStateRunning:
		return receptor.ActualLRPStateRunning
	}

	return ""
}
