package main_test

import (
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Freshness API", func() {
	BeforeEach(func() {
		receptorProcess = ginkgomon.Invoke(receptorRunner)
	})

	AfterEach(func() {
		ginkgomon.Kill(receptorProcess)
	})

	Describe("POST /v1/fresh_domains", func() {
		var postErr error

		BeforeEach(func() {
			freshnessRequest := receptor.FreshDomainBumpRequest{
				Domain:       "domain-0",
				TTLInSeconds: 100,
			}
			postErr = client.BumpFreshDomain(freshnessRequest)
		})

		It("responds without error", func() {
			Ω(postErr).ShouldNot(HaveOccurred())
		})

		It("has the correct data from the bbs", func() {
			freshnesses, err := bbs.Freshnesses()
			Ω(err).ShouldNot(HaveOccurred())

			Ω(freshnesses).Should(HaveLen(1))
			Ω(freshnesses[0].Domain).Should(Equal("domain-0"))
			Ω(freshnesses[0].TTLInSeconds).Should(BeNumerically("<=", 100))
		})
	})

	Describe("GET /v1/fresh_domains", func() {
		var responses []receptor.FreshDomainResponse
		var getErr error

		BeforeEach(func() {
			freshnesses := []models.Freshness{{"domain-0", 100}, {"domain-1", 200}}

			for _, freshness := range freshnesses {
				err := bbs.BumpFreshness(freshness)
				Ω(err).ShouldNot(HaveOccurred())
			}

			responses, getErr = client.FreshDomains()
		})

		It("responds without error", func() {
			Ω(getErr).ShouldNot(HaveOccurred())
		})

		It("has the correct number of responses", func() {
			Ω(responses).Should(HaveLen(2))
		})

		It("has the correct domains from the bbs", func() {
			expectedFreshnesses, err := bbs.Freshnesses()
			Ω(err).ShouldNot(HaveOccurred())

			var expectedDomains []string

			for _, freshness := range expectedFreshnesses {
				expectedDomains = append(expectedDomains, freshness.Domain)
			}

			for _, response := range responses {
				Ω(expectedDomains).Should(ContainElement(response.Domain))
			}
		})

		It("has the correct TTLs from the bbs accounting for degradation", func() {
			expectedFreshnesses, err := bbs.Freshnesses()
			Ω(err).ShouldNot(HaveOccurred())

			expectedDomainsToTTLs := make(map[string]int)

			for _, freshness := range expectedFreshnesses {
				expectedDomainsToTTLs[freshness.Domain] = freshness.TTLInSeconds
			}

			for _, response := range responses {
				expectedTTL, found := expectedDomainsToTTLs[response.Domain]
				Ω(found).Should(BeTrue())
				Ω(expectedTTL).Should(BeNumerically("<=", response.TTLInSeconds))
			}
		})
	})
})
