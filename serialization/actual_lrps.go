package serialization

import (
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/receptor"
	oldmodels "github.com/cloudfoundry-incubator/runtime-schema/models"
)

func ActualLRPProtoToResponse(actualLRP *models.ActualLRP, evacuating bool) receptor.ActualLRPResponse {
	return receptor.ActualLRPResponse{
		ProcessGuid:     actualLRP.ProcessGuid,
		InstanceGuid:    actualLRP.InstanceGuid,
		CellID:          actualLRP.CellId,
		Domain:          actualLRP.Domain,
		Index:           int(actualLRP.Index),
		Address:         actualLRP.Address,
		Ports:           PortMappingFromProto(actualLRP.Ports),
		State:           actualLRPProtoStateToResponseState(actualLRP.State),
		PlacementError:  actualLRP.PlacementError,
		Since:           actualLRP.Since,
		CrashCount:      int(actualLRP.CrashCount),
		CrashReason:     actualLRP.CrashReason,
		Evacuating:      evacuating,
		ModificationTag: actualLRPProtoModificationTagToResponseModificationTag(actualLRP.ModificationTag),
	}
}

func actualLRPProtoStateToResponseState(state string) receptor.ActualLRPState {
	switch state {
	case models.ActualLRPStateUnclaimed:
		return receptor.ActualLRPStateUnclaimed
	case models.ActualLRPStateClaimed:
		return receptor.ActualLRPStateClaimed
	case models.ActualLRPStateRunning:
		return receptor.ActualLRPStateRunning
	case models.ActualLRPStateCrashed:
		return receptor.ActualLRPStateCrashed
	default:
		return receptor.ActualLRPStateInvalid
	}
}

func actualLRPProtoModificationTagToResponseModificationTag(modificationTag *models.ModificationTag) receptor.ModificationTag {
	if modificationTag == nil {
		return receptor.ModificationTag{}
	}
	return receptor.ModificationTag{
		Epoch: modificationTag.Epoch,
		Index: uint(modificationTag.Index),
	}
}

// old code -- delete when BBS server is done

func ActualLRPToResponse(actualLRP oldmodels.ActualLRP, evacuating bool) receptor.ActualLRPResponse {
	return receptor.ActualLRPResponse{
		ProcessGuid:     actualLRP.ProcessGuid,
		InstanceGuid:    actualLRP.InstanceGuid,
		CellID:          actualLRP.CellID,
		Domain:          actualLRP.Domain,
		Index:           actualLRP.Index,
		Address:         actualLRP.Address,
		Ports:           PortMappingFromModel(actualLRP.Ports),
		State:           actualLRPStateToResponseState(actualLRP.State),
		PlacementError:  actualLRP.PlacementError,
		Since:           actualLRP.Since,
		CrashCount:      actualLRP.CrashCount,
		CrashReason:     actualLRP.CrashReason,
		Evacuating:      evacuating,
		ModificationTag: actualLRPModificationTagToResponseModificationTag(actualLRP.ModificationTag),
	}
}

func actualLRPStateToResponseState(state oldmodels.ActualLRPState) receptor.ActualLRPState {
	switch state {
	case models.ActualLRPStateUnclaimed:
		return receptor.ActualLRPStateUnclaimed
	case models.ActualLRPStateClaimed:
		return receptor.ActualLRPStateClaimed
	case models.ActualLRPStateRunning:
		return receptor.ActualLRPStateRunning
	case models.ActualLRPStateCrashed:
		return receptor.ActualLRPStateCrashed
	default:
		return receptor.ActualLRPStateInvalid
	}
}

func actualLRPModificationTagToResponseModificationTag(modificationTag oldmodels.ModificationTag) receptor.ModificationTag {
	return receptor.ModificationTag{
		Epoch: modificationTag.Epoch,
		Index: modificationTag.Index,
	}
}
