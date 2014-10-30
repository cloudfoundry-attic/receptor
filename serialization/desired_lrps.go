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
		EnvironmentVariables: req.EnvironmentVariables,
		Actions:              req.Actions,
		DiskMB:               req.DiskMB,
		MemoryMB:             req.MemoryMB,
		CPUWeight:            req.CPUWeight,
		Ports:                req.Ports,
		Routes:               req.Routes,
		Log:                  req.Log,
	}

	err := lrp.Validate()
	if err != nil {
		return models.DesiredLRP{}, err
	}
	return lrp, nil
}
