package serialization

import (
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

func FreshnessFromRequest(req receptor.FreshDomainBumpRequest) models.Freshness {
	return models.Freshness{
		Domain:       req.Domain,
		TTLInSeconds: req.TTLInSeconds,
	}
}

func FreshnessToResponse(freshness models.Freshness) receptor.FreshDomainResponse {
	return receptor.FreshDomainResponse{
		Domain:       freshness.Domain,
		TTLInSeconds: freshness.TTLInSeconds,
	}
}
