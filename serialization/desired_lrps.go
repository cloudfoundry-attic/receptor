package serialization

import (
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

func DesiredLRPFromRequest(req receptor.CreateDesiredLRPRequest) (models.DesiredLRP, error) {
	lrp := models.DesiredLRP{
		ProcessGuid:          req.ProcessGuid,
		Domain:               req.Domain,
		RootFSPath:           req.RootFSPath,
		Instances:            req.Instances,
		Stack:                req.Stack,
		EnvironmentVariables: EnvironmentVariablesToModel(req.EnvironmentVariables),
		Actions:              req.Actions,
		DiskMB:               req.DiskMB,
		MemoryMB:             req.MemoryMB,
		CPUWeight:            req.CPUWeight,
		Ports:                PortMappingToModel(req.Ports),
		Routes:               req.Routes,
		Log:                  LogConfigToModel(req.Log),
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
		Actions:              lrp.Actions,
		DiskMB:               lrp.DiskMB,
		MemoryMB:             lrp.MemoryMB,
		CPUWeight:            lrp.CPUWeight,
		Ports:                PortMappingFromModel(lrp.Ports),
		Routes:               lrp.Routes,
		Log:                  LogConfigFromModel(lrp.Log),
		Annotation:           lrp.Annotation,
	}
}
