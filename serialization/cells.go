package serialization

import (
	"github.com/cloudfoundry-incubator/locket/presence"
	"github.com/cloudfoundry-incubator/receptor"
)

func CellPresenceToCellResponse(cellPresence presence.CellPresence) receptor.CellResponse {
	return receptor.CellResponse{
		CellID: cellPresence.CellID,
		Zone:   cellPresence.Zone,
		Capacity: receptor.CellCapacity{
			MemoryMB:   cellPresence.Capacity.MemoryMB,
			DiskMB:     cellPresence.Capacity.DiskMB,
			Containers: cellPresence.Capacity.Containers,
		},
		RootFSProviders: cellPresence.RootFSProviders,
	}
}
