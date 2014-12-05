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

func actualLRPStateToResponseState(state models.ActualLRPState) receptor.ActualLRPState {
	switch state {
	case models.ActualLRPStateUnclaimed:
		return receptor.ActualLRPStateUnclaimed
	case models.ActualLRPStateClaimed:
		return receptor.ActualLRPStateClaimed
	case models.ActualLRPStateRunning:
		return receptor.ActualLRPStateRunning
	default:
		return receptor.ActualLRPStateInvalid
	}

	return ""
}
