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
			lrpKey := models.NewActualLRPKey(
				"process-guid-"+index,
				i,
				fmt.Sprintf("domain-%d", i/2),
			)
			containerKey := models.NewActualLRPContainerKey(
				"instance-guid-"+index,
				"cell-id",
			)
			netInfo := models.NewActualLRPNetInfo("the-host", []models.PortMapping{{ContainerPort: 80, HostPort: uint32(1000 + i)}})
			err := bbs.StartActualLRP(lrpKey, containerKey, netInfo, logger)
			Ω(err).ShouldNot(HaveOccurred())
		}
	})

	AfterEach(func() {
		ginkgomon.Kill(receptorProcess)
	})

	Describe("GET /v1/actual_lrps", func() {
		var actualLRPResponses []receptor.ActualLRPResponse
		var getErr error

		BeforeEach(func() {
			actualLRPResponses, getErr = client.ActualLRPs()
		})

		It("responds without an error", func() {
			Ω(getErr).ShouldNot(HaveOccurred())
		})

		It("fetches all of the actual lrps", func() {
			Ω(actualLRPResponses).Should(HaveLen(lrpCount))
		})

		It("has the correct data from the bbs", func() {
			actualLRPs, err := bbs.ActualLRPs()
			Ω(err).ShouldNot(HaveOccurred())

			expectedResponses := make([]receptor.ActualLRPResponse, 0, lrpCount)
			for _, actualLRP := range actualLRPs {
				expectedResponses = append(expectedResponses, serialization.ActualLRPToResponse(actualLRP))
			}

			Ω(actualLRPResponses).Should(ConsistOf(expectedResponses))
		})
	})

	Describe("GET /v1/domains/:domain/actual_lrps", func() {
		var actualLRPResponses []receptor.ActualLRPResponse
		var getErr error

		BeforeEach(func() {
			actualLRPResponses, getErr = client.ActualLRPsByDomain("domain-1")
		})

		It("responds without an error", func() {
			Ω(getErr).ShouldNot(HaveOccurred())
		})

		It("fetches all of the actual lrps", func() {
			Ω(actualLRPResponses).Should(HaveLen(2))
		})

		It("has the correct data from the bbs", func() {
			actualLRPs, err := bbs.ActualLRPsByDomain("domain-1")
			Ω(err).ShouldNot(HaveOccurred())

			expectedResponses := make([]receptor.ActualLRPResponse, 0, 2)
			for _, actualLRP := range actualLRPs {
				expectedResponses = append(expectedResponses, serialization.ActualLRPToResponse(actualLRP))
			}

			Ω(actualLRPResponses).Should(ConsistOf(expectedResponses))
		})
	})

	Describe("GET /v1/actual_lrps/:process_guid", func() {
		var actualLRPResponses []receptor.ActualLRPResponse
		var getErr error

		BeforeEach(func() {
			actualLRPResponses, getErr = client.ActualLRPsByProcessGuid("process-guid-0")
		})

		It("responds without an error", func() {
			Ω(getErr).ShouldNot(HaveOccurred())
		})

		It("fetches all of the actual lrps for the process guid", func() {
			Ω(actualLRPResponses).Should(HaveLen(1))
		})

		It("has the correct data from the bbs", func() {
			actualLRPs, err := bbs.ActualLRPsByProcessGuid("process-guid-0")
			Ω(err).ShouldNot(HaveOccurred())

			expectedResponses := make([]receptor.ActualLRPResponse, 0, 1)
			for _, actualLRP := range actualLRPs {
				expectedResponses = append(expectedResponses, serialization.ActualLRPToResponse(actualLRP))
			}

			Ω(actualLRPResponses).Should(ConsistOf(expectedResponses))
		})
	})

	Describe("GET /v1/actual_lrps/index/:index", func() {
		var actualLRPResponse receptor.ActualLRPResponse
		var getErr error
		var processGuid string
		var index int

		BeforeEach(func() {
			processGuid = "process-guid-0"
			index = 1

			lrpKey := models.NewActualLRPKey(
				processGuid,
				index,
				"domain-0",
			)
			containerKey := models.NewActualLRPContainerKey(
				"instance-guid-0",
				"cell-id",
			)
			netInfo := models.NewActualLRPNetInfo("the-host", []models.PortMapping{{ContainerPort: 80, HostPort: 2345}})
			err := bbs.StartActualLRP(lrpKey, containerKey, netInfo, logger)
			Ω(err).ShouldNot(HaveOccurred())

			actualLRPResponse, getErr = client.ActualLRPByProcessGuidAndIndex(processGuid, index)
			Ω(getErr).ShouldNot(HaveOccurred())
		})

		It("has the correct data from the bbs", func() {
			actualLRP, err := bbs.ActualLRPByProcessGuidAndIndex(processGuid, index)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(actualLRPResponse).Should(Equal(serialization.ActualLRPToResponse(*actualLRP)))
		})
	})
})
