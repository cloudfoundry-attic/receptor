package main_test

import (
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Event", func() {
	var eventSource receptor.EventSource
	var desiredLRP models.DesiredLRP

	JustBeforeEach(func() {
		receptorProcess = ginkgomon.Invoke(receptorRunner)

		var err error
		eventSource, err = client.SubscribeToEvents()
		Ω(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		ginkgomon.Kill(receptorProcess)
	})

	Describe("Desired LRPs", func() {
		BeforeEach(func() {
			desiredLRP = models.DesiredLRP{
				ProcessGuid: "some-guid",
				Domain:      "some-domain",
				Stack:       "some-stack",
				Routes:      []string{"original-route"},
				Action: &models.RunAction{
					Path: "true",
				},
			}
		})

		It("receives events", func() {
			By("creating a DesiredLRP")
			err := bbs.DesireLRP(logger, desiredLRP)
			Ω(err).ShouldNot(HaveOccurred())

			event, err := eventSource.Next()
			Ω(err).ShouldNot(HaveOccurred())

			desiredLRPChangedEvent, ok := event.(receptor.DesiredLRPChangedEvent)
			Ω(ok).Should(BeTrue())
			Ω(desiredLRPChangedEvent.DesiredLRPResponse).Should(Equal(serialization.DesiredLRPToResponse(desiredLRP)))

			By("updating an existing DesiredLRP")
			newRoutes := []string{"new-route"}
			err = bbs.UpdateDesiredLRP(logger, desiredLRP.ProcessGuid, models.DesiredLRPUpdate{Routes: newRoutes})
			Ω(err).ShouldNot(HaveOccurred())

			event, err = eventSource.Next()
			Ω(err).ShouldNot(HaveOccurred())

			desiredLRPChangedEvent, ok = event.(receptor.DesiredLRPChangedEvent)
			Ω(ok).Should(BeTrue())
			Ω(desiredLRPChangedEvent.DesiredLRPResponse.Routes).Should(Equal(newRoutes))

			By("removing the DesiredLRP")
			err = bbs.RemoveDesiredLRPByProcessGuid(logger, desiredLRP.ProcessGuid)
			Ω(err).ShouldNot(HaveOccurred())

			event, err = eventSource.Next()
			Ω(err).ShouldNot(HaveOccurred())

			desiredLRPRemovedEvent, ok := event.(receptor.DesiredLRPRemovedEvent)
			Ω(ok).Should(BeTrue())
			Ω(desiredLRPRemovedEvent.DesiredLRPResponse.ProcessGuid).Should(Equal(desiredLRP.ProcessGuid))
		})
	})

	Describe("Actual LRPs", func() {
		BeforeEach(func() {
			desiredLRP = models.DesiredLRP{
				ProcessGuid: "some-guid",
				Domain:      "some-domain",
				Stack:       "some-stack",
				Instances:   1,
				Action: &models.RunAction{
					Path: "true",
				},
			}
		})

		It("receives events", func() {
			By("creating a ActualLRP")
			err := bbs.CreateActualLRP(desiredLRP, 0, logger)
			Ω(err).ShouldNot(HaveOccurred())

			actualLRP, err := bbs.ActualLRPByProcessGuidAndIndex(desiredLRP.ProcessGuid, 0)
			Ω(err).ShouldNot(HaveOccurred())

			event, err := eventSource.Next()
			Ω(err).ShouldNot(HaveOccurred())

			actualLRPChangedEvent, ok := event.(receptor.ActualLRPChangedEvent)
			Ω(ok).Should(BeTrue())
			Ω(actualLRPChangedEvent.ActualLRPResponse).Should(Equal(serialization.ActualLRPToResponse(actualLRP)))

			By("updating an existing ActualLRP")
			err = bbs.ClaimActualLRP(actualLRP.ActualLRPKey, models.NewActualLRPContainerKey("some-instance-guid", "some-cell-id"), logger)
			Ω(err).ShouldNot(HaveOccurred())

			actualLRP, err = bbs.ActualLRPByProcessGuidAndIndex(desiredLRP.ProcessGuid, 0)
			Ω(err).ShouldNot(HaveOccurred())

			event, err = eventSource.Next()
			Ω(err).ShouldNot(HaveOccurred())

			actualLRPChangedEvent, ok = event.(receptor.ActualLRPChangedEvent)
			Ω(ok).Should(BeTrue())
			Ω(actualLRPChangedEvent.ActualLRPResponse).Should(Equal(serialization.ActualLRPToResponse(actualLRP)))

			By("removing the ActualLRP")
			err = bbs.RemoveActualLRP(actualLRP.ActualLRPKey, actualLRP.ActualLRPContainerKey, logger)
			Ω(err).ShouldNot(HaveOccurred())

			event, err = eventSource.Next()
			Ω(err).ShouldNot(HaveOccurred())

			actualLRPRemovedEvent, ok := event.(receptor.ActualLRPRemovedEvent)
			Ω(ok).Should(BeTrue())
			Ω(actualLRPRemovedEvent.ActualLRPResponse).Should(Equal(serialization.ActualLRPToResponse(actualLRP)))
		})
	})
})
