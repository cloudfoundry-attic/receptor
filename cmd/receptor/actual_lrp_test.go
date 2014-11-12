package main_test

import (
	"strconv"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Actual LRP API", func() {

	BeforeEach(func() {
		receptorProcess = ginkgomon.Invoke(receptorRunner)
	})

	AfterEach(func() {
		ginkgomon.Kill(receptorProcess)
	})

	Describe("GET /actual_lrps", func() {
		var actualLRPResponses []receptor.ActualLRPResponse
		var getErr error

		const expectedLRPCount = 6 // fingers - right hand

		BeforeEach(func() {
			for i := 0; i < expectedLRPCount; i++ {
				index := strconv.Itoa(i)
				lrp, err := models.NewActualLRP(
					"process-guid-"+index,
					"instance-guid-"+index,
					"executor-id",
					"domain",
					i,
				)
				Ω(err).ShouldNot(HaveOccurred())

				err = bbs.ReportActualLRPAsRunning(lrp, "executor-id")
				Ω(err).ShouldNot(HaveOccurred())
			}

			actualLRPResponses, getErr = client.GetAllActualLRPs()
		})

		It("responds without an error", func() {
			Ω(getErr).ShouldNot(HaveOccurred())
		})

		It("fetches all of the actual lrps", func() {
			Ω(actualLRPResponses).Should(HaveLen(expectedLRPCount))
		})

		It("has the correct data from the bbs", func() {
			actualLRPs, err := bbs.GetAllActualLRPs()
			Ω(err).ShouldNot(HaveOccurred())

			expectedResponses := make([]receptor.ActualLRPResponse, 0, expectedLRPCount)
			for _, actualLRP := range actualLRPs {
				expectedResponses = append(expectedResponses, serialization.ActualLRPToResponse(actualLRP))
			}

			Ω(actualLRPResponses).Should(ConsistOf(expectedResponses))
		})
	})

})
