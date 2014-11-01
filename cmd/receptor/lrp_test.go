package main_test

import (
	"fmt"
	"sync/atomic"

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
		var lrpToCreate receptor.CreateDesiredLRPRequest
		var createErr error

		BeforeEach(func() {
			lrpToCreate = newValidCreateDesiredLRPRequest()
			createErr = client.CreateDesiredLRP(lrpToCreate)
		})

		It("responds without an error", func() {
			Ω(createErr).ShouldNot(HaveOccurred())
		})

		It("desires an LRP in the BBS", func() {
			Eventually(bbs.GetAllDesiredLRPs).Should(HaveLen(1))
			desiredLRPs, err := bbs.GetAllDesiredLRPs()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(desiredLRPs[0].ProcessGuid).To(Equal(lrpToCreate.ProcessGuid))
		})
	})

	Describe("GET /desired_lrps", func() {
		var lrpRequests []receptor.CreateDesiredLRPRequest
		var lrpResponses []receptor.DesiredLRPResponse
		const expectedLRPcount = 6
		var getErr error

		BeforeEach(func() {
			lrpRequests = make([]receptor.CreateDesiredLRPRequest, expectedLRPcount)
			for i := 0; i < expectedLRPcount; i++ {
				lrpRequests[i] = newValidCreateDesiredLRPRequest()
				err := client.CreateDesiredLRP(lrpRequests[i])
				Ω(err).ShouldNot(HaveOccurred())
			}
			lrpResponses, getErr = client.GetAllDesiredLRPs()
		})

		It("responds without an error", func() {
			Ω(getErr).ShouldNot(HaveOccurred())
		})

		It("fetches all of the desired lrps", func() {
			Ω(lrpResponses).Should(HaveLen(expectedLRPcount))
		})
	})
})

var processId int64

func newValidCreateDesiredLRPRequest() receptor.CreateDesiredLRPRequest {
	atomic.AddInt64(&processId, 1)

	return receptor.CreateDesiredLRPRequest{
		ProcessGuid: fmt.Sprintf("process-guid-%d", processId),
		Domain:      "test-domain",
		Stack:       "some-stack",
		Instances:   1,
		Actions: []models.ExecutorAction{
			{
				models.RunAction{
					Path: "/bin/bash",
				},
			},
		},
	}
}
