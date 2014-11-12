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

	Describe("GET /desired_lrps/:process_guid/actual_lrps?index=:index", func() {
		var actualLRPResponses []receptor.ActualLRPResponse
		var getErr error

		BeforeEach(func() {
			lrp, err := models.NewActualLRP(
				"process-guid-0",
				"instance-guid-0",
				"executor-id",
				"domain-0",
				1,
			)
			Ω(err).ShouldNot(HaveOccurred())
			err = bbs.ReportActualLRPAsRunning(lrp, "executor-id")
			Ω(err).ShouldNot(HaveOccurred())
			actualLRPResponses, getErr = client.GetAllActualLRPsByProcessGuidAndIndex("process-guid-0", 0)
		})

		It("responds without an error", func() {
			Ω(getErr).ShouldNot(HaveOccurred())
		})

		It("fetches all of the actual lrps for the process guid", func() {
			Ω(actualLRPResponses).Should(HaveLen(1))
		})

		It("has the correct data from the bbs", func() {
			actualLRPs, err := bbs.GetActualLRPsByProcessGuidAndIndex("process-guid-0", 0)
			Ω(err).ShouldNot(HaveOccurred())

			expectedResponses := make([]receptor.ActualLRPResponse, 0, 1)
			for _, actualLRP := range actualLRPs {
				expectedResponses = append(expectedResponses, serialization.ActualLRPToResponse(actualLRP)) // Fix to only include guids with index 0
			}

			Ω(actualLRPResponses).Should(ConsistOf(expectedResponses))
		})
	})

	Describe("DELETE /desired_lrps/:process_guid/actual_lrps?index=:index", func() {
		var killErr error

		BeforeEach(func() {
			lrp, err := models.NewActualLRP(
				"process-guid-0",
				"instance-guid-0",
				"executor-id",
				"domain-0",
				1,
			)
			Ω(err).ShouldNot(HaveOccurred())
			err = bbs.ReportActualLRPAsRunning(lrp, "executor-id")
			Ω(err).ShouldNot(HaveOccurred())

			killErr = client.KillActualLRPsByProcessGuidAndIndex("process-guid-0", 0)
		})

		It("responds without an error", func() {
			Ω(killErr).ShouldNot(HaveOccurred())
		})

		It("places the correct stop instance requests in the bbs", func() {
			stopLRPInstances, err := bbs.GetAllStopLRPInstances()
			Ω(err).ShouldNot(HaveOccurred())

			Ω(stopLRPInstances).Should(HaveLen(1))
			Ω(stopLRPInstances[0].ProcessGuid).Should(Equal("process-guid-0"))
			Ω(stopLRPInstances[0].Index).Should(Equal(0))
		})
	})
})
