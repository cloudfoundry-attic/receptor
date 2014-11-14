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
		var request receptor.FreshDomainBumpRequest
		var freshness models.Freshness

		BeforeEach(func() {
			request = receptor.FreshDomainBumpRequest{
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

	Describe("FreshnessToResponse", func() {
		var freshness models.Freshness
		var response receptor.FreshDomainResponse

		BeforeEach(func() {
			freshness = models.Freshness{
				Domain:       "the-domain",
				TTLInSeconds: 100,
			}
			response = serialization.FreshnessToResponse(freshness)
		})

		It("serializes the Freshness model into a response, preserving attributes", func() {
			Ω(response.Domain).Should(Equal("the-domain"))
			Ω(response.TTLInSeconds).Should(Equal(100))
		})
	})
})
