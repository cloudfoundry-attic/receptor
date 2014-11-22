package serialization

import (
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

func DesiredLRPFromRequest(req receptor.DesiredLRPCreateRequest) models.DesiredLRP {
	return models.DesiredLRP{
		ProcessGuid:          req.ProcessGuid,
		Domain:               req.Domain,
		RootFSPath:           req.RootFSPath,
		Instances:            req.Instances,
		Stack:                req.Stack,
		EnvironmentVariables: EnvironmentVariablesToModel(req.EnvironmentVariables),
		Setup:                req.Setup,
		Action:               req.Action,
		Monitor:              req.Monitor,
		StartTimeout:         req.StartTimeout,
		DiskMB:               req.DiskMB,
		MemoryMB:             req.MemoryMB,
		CPUWeight:            req.CPUWeight,
		Ports:                req.Ports,
		Routes:               req.Routes,
		LogGuid:              req.LogGuid,
		LogSource:            req.LogSource,
		Annotation:           req.Annotation,
	}
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
		Action:               lrp.Action,
		Monitor:              lrp.Monitor,
		StartTimeout:         lrp.StartTimeout,
		DiskMB:               lrp.DiskMB,
		MemoryMB:             lrp.MemoryMB,
		CPUWeight:            lrp.CPUWeight,
		Ports:                lrp.Ports,
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
