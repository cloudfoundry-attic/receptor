package serialization

import (
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

func DesiredLRPFromRequest(req receptor.DesiredLRPCreateRequest) (models.DesiredLRP, error) {
	var action models.ExecutorAction
	if req.Action != nil {
		action = *req.Action
	}

	lrp := models.DesiredLRP{
		ProcessGuid:          req.ProcessGuid,
		Domain:               req.Domain,
		RootFSPath:           req.RootFSPath,
		Instances:            req.Instances,
		Stack:                req.Stack,
		EnvironmentVariables: EnvironmentVariablesToModel(req.EnvironmentVariables),
		Setup:                req.Setup,
		Action:               action,
		Monitor:              req.Monitor,
		DiskMB:               req.DiskMB,
		MemoryMB:             req.MemoryMB,
		CPUWeight:            req.CPUWeight,
		Ports:                PortMappingToModel(req.Ports),
		Routes:               req.Routes,
		LogGuid:              req.LogGuid,
		LogSource:            req.LogSource,
		Annotation:           req.Annotation,
	}

	err := lrp.Validate()
	if err != nil {
		return models.DesiredLRP{}, err
	}
	return lrp, nil
}

func DesiredLRPToResponse(lrp models.DesiredLRP) receptor.DesiredLRPResponse {
	return receptor.DesiredLRPResponse{
		ProcessGuid:          lrp.ProcessGuid,
		Domain:               lrp.Domain,
		RootFSPath:           lrp.RootFSPath,
		Instances:            lrp.Instances,
		Stack:                lrp.Stack,
		EnvironmentVariables: EnvironmentVariablesFromModel(lrp.EnvironmentVariables),
		Setup:                lrp.Setup,
		Action:               &lrp.Action,
		Monitor:              lrp.Monitor,
		DiskMB:               lrp.DiskMB,
		MemoryMB:             lrp.MemoryMB,
		CPUWeight:            lrp.CPUWeight,
		Ports:                PortMappingFromModel(lrp.Ports),
		Routes:               lrp.Routes,
		LogGuid:              lrp.LogGuid,
		LogSource:            lrp.LogSource,
		Annotation:           lrp.Annotation,
	}
}

func DesiredLRPUpdateFromRequest(req receptor.DesiredLRPUpdateRequest) models.DesiredLRPUpdate {
	return models.DesiredLRPUpdate{
		Instances:  req.Instances,
		Routes:     req.Routes,
		Annotation: req.Annotation,
	}
}
