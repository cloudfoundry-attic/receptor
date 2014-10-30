package main_test

import (
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Desired LRP API", func() {

	BeforeEach(func() {
		receptorProcess = ginkgomon.Invoke(receptorRunner)
	})

	AfterEach(func() {
		ginkgomon.Kill(receptorProcess)
	})

	Describe("POST /desired_lrps", func() {
		var (
			lrpToCreate receptor.CreateDesiredLRPRequest
			err         error
		)

		BeforeEach(func() {
			lrpToCreate = receptor.CreateDesiredLRPRequest{
				ProcessGuid: "process-guid-1",
				Domain:      "test-domain",
				Stack:       "some-stack",
				Actions: []models.ExecutorAction{
					{
						models.RunAction{
							Path: "/bin/bash",
						},
					},
				},
			}

			err = client.CreateDesiredLRP(lrpToCreate)
		})

		It("responds without an error", func() {
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("desires an LRP in the BBS", func() {
			Eventually(bbs.GetAllDesiredLRPs).Should(HaveLen(1))
			desiredLRPs, err := bbs.GetAllDesiredLRPs()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(desiredLRPs[0].ProcessGuid).To(Equal("process-guid-1"))
		})
	})
})
