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
