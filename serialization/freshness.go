package serialization

import (
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

func FreshnessFromRequest(req receptor.FreshDomainCreateRequest) models.Freshness {
	return models.Freshness{
		Domain:       req.Domain,
		TTLInSeconds: req.TTLInSeconds,
	}
}
