package serialization_test

import (
	"github.com/cloudfoundry-incubator/receptor"
	. "github.com/cloudfoundry-incubator/receptor/serialization"
	"github.com/cloudfoundry-incubator/runtime-schema/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DesiredLRP Serialization", func() {
	Describe("DesiredLRPFromRequest", func() {
		var request receptor.CreateDesiredLRPRequest
		var desiredLRP models.DesiredLRP

		BeforeEach(func() {
			request = receptor.CreateDesiredLRPRequest{
				ProcessGuid: "the-process-guid",
				Domain:      "the-domain",
				Stack:       "the-stack",
				RootFSPath:  "the-rootfs-path",
				Annotation:  "foo",
				Instances:   1,
				Actions: []models.ExecutorAction{
					{
						Action: &models.RunAction{
							Path: "the-path",
						},
					},
				},
			}
		})
		JustBeforeEach(func() {
			var err error
			desiredLRP, err = DesiredLRPFromRequest(request)
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("translates the request into a DesiredLRP model, preserving attributes", func() {
			Ω(desiredLRP.ProcessGuid).Should(Equal("the-process-guid"))
			Ω(desiredLRP.Domain).Should(Equal("the-domain"))
			Ω(desiredLRP.Stack).Should(Equal("the-stack"))
			Ω(desiredLRP.RootFSPath).Should(Equal("the-rootfs-path"))
			Ω(desiredLRP.Annotation).Should(Equal("foo"))
		})
	})
})
