package serialization_test

import (
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	"github.com/cloudfoundry-incubator/runtime-schema/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Freshness Serialization", func() {
	Describe("FreshnessFromRequest", func() {
		var request receptor.FreshDomainCreateRequest
		var freshness models.Freshness

		BeforeEach(func() {
			request = receptor.FreshDomainCreateRequest{
				Domain:       "the-domain",
				TTLInSeconds: 100,
			}
			freshness = serialization.FreshnessFromRequest(request)
		})

		It("translates the request into a Freshness model, preserving attributes", func() {
			Ω(freshness.Domain).Should(Equal("the-domain"))
			Ω(freshness.TTLInSeconds).Should(Equal(100))
		})
	})
})
