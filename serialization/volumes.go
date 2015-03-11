package serialization

import (
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

func VolumeSetFromCreateRequest(volReq receptor.VolumeSetCreateRequest) models.VolumeSet {
	return models.VolumeSet{
		VolumeSetGuid:    volReq.VolumeSetGuid,
		Stack:            volReq.Stack,
		Instances:        volReq.Instances,
		SizeMB:           volReq.SizeMB,
		ReservedMemoryMB: volReq.ReservedMemoryMB,
	}
}

func VolumeToResponse(volume models.Volume) receptor.VolumeResponse {
	return receptor.VolumeResponse{
		VolumeSetGuid:    volume.VolumeSetGuid,
		VolumeGuid:       volume.VolumeGuid,
		CellID:           volume.CellID,
		Index:            volume.Index,
		SizeMB:           volume.SizeMB,
		ReservedMemoryMB: volume.ReservedMemoryMB,
		State:            volumeStateForResponse(volume.State),
		PlacementError:   volume.PlacementError,
		Since:            volume.Since,
	}
}

func volumeStateForResponse(state models.VolumeState) receptor.VolumeState {
	switch state {
	case models.VolumeStatePending:
		return receptor.VolumeStatePending
	case models.VolumeStateRunning:
		return receptor.VolumeStateRunning
	case models.VolumeStateFailed:
		return receptor.VolumeStateFailed
	default:
		return ""
	}
}
