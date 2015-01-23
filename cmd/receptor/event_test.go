package main_test

import (
	"time"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Event", func() {
	var eventSource receptor.EventSource
	var events chan receptor.Event
	var done chan struct{}
	var desiredLRP models.DesiredLRP

	JustBeforeEach(func() {
		receptorProcess = ginkgomon.Invoke(receptorRunner)

		var err error
		eventSource, err = client.SubscribeToEvents()
		Ω(err).ShouldNot(HaveOccurred())

		events = make(chan receptor.Event)
		done = make(chan struct{})

		go func() {
			defer close(done)
			for {
				event, err := eventSource.Next()
				if err != nil {
					close(events)
					return
				}
				events <- event
			}
		}()

		primerLRP := models.DesiredLRP{
			ProcessGuid: "primer-guid",
			Domain:      "primer-domain",
			Stack:       "primer-stack",
			Routes:      []string{"primer-route"},
			Action: &models.RunAction{
				Path: "true",
			},
		}

		err = bbs.DesireLRP(logger, primerLRP)
		Ω(err).ShouldNot(HaveOccurred())

	PRIMING:
		for {
			select {
			case <-events:
				break PRIMING
			case <-time.After(50 * time.Millisecond):
				err = bbs.UpdateDesiredLRP(logger, primerLRP.ProcessGuid, models.DesiredLRPUpdate{Routes: []string{"garbage-route"}})
				Ω(err).ShouldNot(HaveOccurred())
			}
		}

		err = bbs.RemoveDesiredLRPByProcessGuid(logger, primerLRP.ProcessGuid)
		Ω(err).ShouldNot(HaveOccurred())

		var event receptor.Event
		for {
			Eventually(events).Should(Receive(&event))
			if event.EventType() == receptor.EventTypeDesiredLRPRemoved {
				break
			}
		}
	})

	AfterEach(func() {
		ginkgomon.Kill(receptorProcess)
		err := eventSource.Close()
		Ω(err).ShouldNot(HaveOccurred())
		Eventually(done).Should(BeClosed())
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

			var event receptor.Event
			Eventually(events).Should(Receive(&event))

			desiredLRPCreatedEvent, ok := event.(receptor.DesiredLRPCreatedEvent)
			Ω(ok).Should(BeTrue())
			Ω(desiredLRPCreatedEvent.DesiredLRPResponse).Should(Equal(serialization.DesiredLRPToResponse(desiredLRP)))

			By("updating an existing DesiredLRP")
			newRoutes := []string{"new-route"}
			err = bbs.UpdateDesiredLRP(logger, desiredLRP.ProcessGuid, models.DesiredLRPUpdate{Routes: newRoutes})
			Ω(err).ShouldNot(HaveOccurred())

			Eventually(events).Should(Receive(&event))

			desiredLRPChangedEvent, ok := event.(receptor.DesiredLRPChangedEvent)
			Ω(ok).Should(BeTrue())
			Ω(desiredLRPChangedEvent.After.Routes).Should(Equal(newRoutes))

			By("removing the DesiredLRP")
			err = bbs.RemoveDesiredLRPByProcessGuid(logger, desiredLRP.ProcessGuid)
			Ω(err).ShouldNot(HaveOccurred())

			Eventually(events).Should(Receive(&event))

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

			var event receptor.Event

			Eventually(events).Should(Receive(&event))

			actualLRPCreatedEvent, ok := event.(receptor.ActualLRPCreatedEvent)
			Ω(ok).Should(BeTrue())
			Ω(actualLRPCreatedEvent.ActualLRPResponse).Should(Equal(serialization.ActualLRPToResponse(actualLRP)))

			By("updating an existing ActualLRP")
			err = bbs.ClaimActualLRP(actualLRP.ActualLRPKey, models.NewActualLRPContainerKey("some-instance-guid", "some-cell-id"), logger)
			Ω(err).ShouldNot(HaveOccurred())

			actualLRP, err = bbs.ActualLRPByProcessGuidAndIndex(desiredLRP.ProcessGuid, 0)
			Ω(err).ShouldNot(HaveOccurred())

			Eventually(events).Should(Receive(&event))

			actualLRPChangedEvent, ok := event.(receptor.ActualLRPChangedEvent)
			Ω(ok).Should(BeTrue())
			Ω(actualLRPChangedEvent.After).Should(Equal(serialization.ActualLRPToResponse(actualLRP)))

			By("removing the ActualLRP")
			err = bbs.RemoveActualLRP(actualLRP.ActualLRPKey, actualLRP.ActualLRPContainerKey, logger)
			Ω(err).ShouldNot(HaveOccurred())

			Eventually(events).Should(Receive(&event))

			actualLRPRemovedEvent, ok := event.(receptor.ActualLRPRemovedEvent)
			Ω(ok).Should(BeTrue())
			Ω(actualLRPRemovedEvent.ActualLRPResponse).Should(Equal(serialization.ActualLRPToResponse(actualLRP)))
		})
	})
})
