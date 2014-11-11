package main_test

import (
	"fmt"
	"strconv"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Actual LRP API", func() {
	const lrpCount = 6

	BeforeEach(func() {
		receptorProcess = ginkgomon.Invoke(receptorRunner)

		for i := 0; i < lrpCount; i++ {
			index := strconv.Itoa(i)
			lrp, err := models.NewActualLRP(
				"process-guid-"+index,
				"instance-guid-"+index,
				"executor-id",
				fmt.Sprintf("domain-%d", i/2),
				i,
				models.ActualLRPStateRunning,
				99999999999,
			)
			Ω(err).ShouldNot(HaveOccurred())
			err = bbs.ReportActualLRPAsRunning(lrp, "executor-id")
			Ω(err).ShouldNot(HaveOccurred())
		}
	})

	AfterEach(func() {
		ginkgomon.Kill(receptorProcess)
	})

	Describe("GET /actual_lrps", func() {
		var actualLRPResponses []receptor.ActualLRPResponse
		var getErr error

		BeforeEach(func() {
			actualLRPResponses, getErr = client.GetAllActualLRPs()
		})

		It("responds without an error", func() {
			Ω(getErr).ShouldNot(HaveOccurred())
		})

		It("fetches all of the actual lrps", func() {
			Ω(actualLRPResponses).Should(HaveLen(lrpCount))
		})

		It("has the correct data from the bbs", func() {
			actualLRPs, err := bbs.GetAllActualLRPs()
			Ω(err).ShouldNot(HaveOccurred())

			expectedResponses := make([]receptor.ActualLRPResponse, 0, lrpCount)
			for _, actualLRP := range actualLRPs {
				expectedResponses = append(expectedResponses, serialization.ActualLRPToResponse(actualLRP))
			}

			Ω(actualLRPResponses).Should(ConsistOf(expectedResponses))
		})
	})

	Describe("GET /domains/:domain/actual_lrps", func() {
		var actualLRPResponses []receptor.ActualLRPResponse
		var getErr error

		BeforeEach(func() {
			actualLRPResponses, getErr = client.GetAllActualLRPsByDomain("domain-1")
		})

		It("responds without an error", func() {
			Ω(getErr).ShouldNot(HaveOccurred())
		})

		It("fetches all of the actual lrps", func() {
			Ω(actualLRPResponses).Should(HaveLen(2))
		})

		It("has the correct data from the bbs", func() {
			actualLRPs, err := bbs.GetAllActualLRPsByDomain("domain-1")
			Ω(err).ShouldNot(HaveOccurred())

			expectedResponses := make([]receptor.ActualLRPResponse, 0, 2)
			for _, actualLRP := range actualLRPs {
				expectedResponses = append(expectedResponses, serialization.ActualLRPToResponse(actualLRP))
			}

			Ω(actualLRPResponses).Should(ConsistOf(expectedResponses))
		})
	})

	Describe("GET /desired_lrps/:process_guid/actual_lrps", func() {
		var actualLRPResponses []receptor.ActualLRPResponse
		var getErr error

		BeforeEach(func() {
			actualLRPResponses, getErr = client.GetAllActualLRPsByProcessGuid("process-guid-0")
		})

		It("responds without an error", func() {
			Ω(getErr).ShouldNot(HaveOccurred())
		})

		It("fetches all of the actual lrps for the process guid", func() {
			Ω(actualLRPResponses).Should(HaveLen(1))
		})

		It("has the correct data from the bbs", func() {
			actualLRPs, err := bbs.GetActualLRPsByProcessGuid("process-guid-0")
			Ω(err).ShouldNot(HaveOccurred())

			expectedResponses := make([]receptor.ActualLRPResponse, 0, 1)
			for _, actualLRP := range actualLRPs {
				expectedResponses = append(expectedResponses, serialization.ActualLRPToResponse(actualLRP))
			}

			Ω(actualLRPResponses).Should(ConsistOf(expectedResponses))
		})
	})
})
