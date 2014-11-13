package main_test

import (
	"github.com/cloudfoundry-incubator/receptor"
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

	Describe("POST /fresh_domains", func() {
		var postErr error

		BeforeEach(func() {
			freshnessRequest := receptor.FreshDomainCreateRequest{
				Domain:       "domain-0",
				TTLInSeconds: 100,
			}
			postErr = client.CreateFreshDomain(freshnessRequest)
		})

		It("responds without error", func() {
			Ω(postErr).ShouldNot(HaveOccurred())
		})

		It("has the correct data from the bbs", func() {
			freshnesses, err := bbs.GetAllFreshness()
			Ω(err).ShouldNot(HaveOccurred())

			Ω(freshnesses).Should(HaveLen(1))
			Ω(freshnesses[0]).Should(Equal("domain-0"))
			// TODO: check that TTL <= 100 once BBS provides it in Freshness response
		})
	})
})
