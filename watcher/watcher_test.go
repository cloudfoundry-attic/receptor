package watcher_test

import (
	"errors"
	"os"
	"time"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/event/eventfakes"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	"github.com/cloudfoundry-incubator/receptor/watcher"
	"github.com/cloudfoundry-incubator/runtime-schema/bbs/fake_bbs"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/cloudfoundry/gunk/timeprovider/faketimeprovider"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager/lagertest"
	"github.com/tedsuo/ifrit"
)

var _ = Describe("Watcher", func() {
	const (
		expectedProcessGuid  = "some-process-guid"
		expectedInstanceGuid = "some-instance-guid"
		retryWaitDuration    = 50 * time.Millisecond
	)

	var (
		bbs             *fake_bbs.FakeReceptorBBS
		hub             *eventfakes.FakeHub
		timeProvider    *faketimeprovider.FakeTimeProvider
		receptorWatcher watcher.Watcher
		process         ifrit.Process

		desiredLRPCreates chan models.DesiredLRP
		desiredLRPUpdates chan models.DesiredLRPChange
		desiredLRPDeletes chan models.DesiredLRP
		desiredLRPErrors  chan error

		actualLRPCreates chan models.ActualLRP
		actualLRPUpdates chan models.ActualLRPChange
		actualLRPDeletes chan models.ActualLRP
		actualLRPErrors  chan error
	)

	BeforeEach(func() {
		bbs = new(fake_bbs.FakeReceptorBBS)
		hub = new(eventfakes.FakeHub)
		timeProvider = faketimeprovider.New(time.Now())
		logger := lagertest.NewTestLogger("test")

		desiredLRPCreates = make(chan models.DesiredLRP)
		desiredLRPUpdates = make(chan models.DesiredLRPChange)
		desiredLRPDeletes = make(chan models.DesiredLRP)
		desiredLRPErrors = make(chan error)

		actualLRPCreates = make(chan models.ActualLRP)
		actualLRPUpdates = make(chan models.ActualLRPChange)
		actualLRPDeletes = make(chan models.ActualLRP)
		actualLRPErrors = make(chan error)

		bbs.WatchForDesiredLRPChangesReturns(desiredLRPCreates, desiredLRPUpdates, desiredLRPDeletes, desiredLRPErrors)
		bbs.WatchForActualLRPChangesReturns(actualLRPCreates, actualLRPUpdates, actualLRPDeletes, actualLRPErrors)

		receptorWatcher = watcher.NewWatcher(bbs, hub, timeProvider, retryWaitDuration, logger)
		process = ifrit.Invoke(receptorWatcher)
	})

	AfterEach(func() {
		process.Signal(os.Interrupt)
		Eventually(process.Wait()).Should(Receive())
	})

	Describe("Desired LRP changes", func() {
		var desiredLRP models.DesiredLRP

		BeforeEach(func() {
			desiredLRP = models.DesiredLRP{
				Action: &models.RunAction{
					Path: "ls",
				},
				Domain:      "tests",
				ProcessGuid: expectedProcessGuid,
			}
		})

		Context("when a create arrives", func() {
			BeforeEach(func() {
				desiredLRPCreates <- desiredLRP
			})

			It("emits a DesiredLRPCreatedEvent to the hub", func() {
				Eventually(hub.EmitCallCount).Should(Equal(1))
				event := hub.EmitArgsForCall(0)

				desiredLRPCreatedEvent, ok := event.(receptor.DesiredLRPCreatedEvent)
				Ω(ok).Should(BeTrue())
				Ω(desiredLRPCreatedEvent.DesiredLRPResponse).Should(Equal(serialization.DesiredLRPToResponse(desiredLRP)))
			})
		})

		Context("when a change arrives", func() {
			BeforeEach(func() {
				desiredLRPUpdates <- models.DesiredLRPChange{Before: desiredLRP, After: desiredLRP}
			})

			It("emits a DesiredLRPChangedEvent to the hub", func() {
				Eventually(hub.EmitCallCount).Should(Equal(1))
				event := hub.EmitArgsForCall(0)

				desiredLRPChangedEvent, ok := event.(receptor.DesiredLRPChangedEvent)
				Ω(ok).Should(BeTrue())
				Ω(desiredLRPChangedEvent.Before).Should(Equal(serialization.DesiredLRPToResponse(desiredLRP)))
				Ω(desiredLRPChangedEvent.After).Should(Equal(serialization.DesiredLRPToResponse(desiredLRP)))
			})
		})

		Context("when a delete arrives", func() {
			BeforeEach(func() {
				desiredLRPDeletes <- desiredLRP
			})

			It("emits a DesiredLRPRemovedEvent to the hub", func() {
				Eventually(hub.EmitCallCount).Should(Equal(1))
				event := hub.EmitArgsForCall(0)

				desiredLRPRemovedEvent, ok := event.(receptor.DesiredLRPRemovedEvent)
				Ω(ok).Should(BeTrue())
				Ω(desiredLRPRemovedEvent.DesiredLRPResponse).Should(Equal(serialization.DesiredLRPToResponse(desiredLRP)))
			})
		})

		Context("when watching for change fails", func() {
			BeforeEach(func() {
				desiredLRPErrors <- errors.New("bbs watch failed")

				// avoid issues with race detector when the next test's
				// BeforeEach resets the changes channel
				changeChan := desiredLRPCreates
				desiredLRPToSend := desiredLRP
				go func() { changeChan <- desiredLRPToSend }()
			})

			It("should retry after the wait duration", func() {
				timeProvider.Increment(retryWaitDuration / 2)
				Consistently(hub.EmitCallCount).Should(BeZero())
				timeProvider.Increment(retryWaitDuration * 2)
				Eventually(hub.EmitCallCount).Should(Equal(1))
			})

			It("should be possible to SIGINT the route emitter", func() {
				process.Signal(os.Interrupt)
				Eventually(process.Wait()).Should(Receive())
			})
		})
	})

	Describe("Actual LRP changes", func() {
		var actualLRP models.ActualLRP

		BeforeEach(func() {
			actualLRP = models.ActualLRP{
				ActualLRPKey:          models.NewActualLRPKey(expectedProcessGuid, 1, "domain"),
				ActualLRPContainerKey: models.NewActualLRPContainerKey(expectedInstanceGuid, "cell-id"),
			}
		})

		Context("when a create arrives", func() {
			BeforeEach(func() {
				actualLRPCreates <- actualLRP
			})

			It("emits an ActualLRPCreatedEvent to the hub", func() {
				Eventually(hub.EmitCallCount).Should(Equal(1))
				event := hub.EmitArgsForCall(0)

				actualLRPCreatedEvent, ok := event.(receptor.ActualLRPCreatedEvent)
				Ω(ok).Should(BeTrue())
				Ω(actualLRPCreatedEvent.ActualLRPResponse).Should(Equal(serialization.ActualLRPToResponse(actualLRP)))
			})
		})

		Context("when a change arrives", func() {
			BeforeEach(func() {
				actualLRPUpdates <- models.ActualLRPChange{Before: actualLRP, After: actualLRP}
			})

			It("emits an ActualLRPChangedEvent to the hub", func() {
				Eventually(hub.EmitCallCount).Should(Equal(1))
				event := hub.EmitArgsForCall(0)

				actualLRPChangedEvent, ok := event.(receptor.ActualLRPChangedEvent)
				Ω(ok).Should(BeTrue())
				Ω(actualLRPChangedEvent.Before).Should(Equal(serialization.ActualLRPToResponse(actualLRP)))
				Ω(actualLRPChangedEvent.After).Should(Equal(serialization.ActualLRPToResponse(actualLRP)))
			})
		})

		Context("when a delete arrives", func() {
			BeforeEach(func() {
				actualLRPDeletes <- actualLRP
			})

			It("emits an ActualLRPRemovedEvent to the hub", func() {
				Eventually(hub.EmitCallCount).Should(Equal(1))
				event := hub.EmitArgsForCall(0)

				actualLRPRemovedEvent, ok := event.(receptor.ActualLRPRemovedEvent)
				Ω(ok).Should(BeTrue())
				Ω(actualLRPRemovedEvent.ActualLRPResponse).Should(Equal(serialization.ActualLRPToResponse(actualLRP)))
			})
		})

		Context("when watching for change fails", func() {
			BeforeEach(func() {
				actualLRPErrors <- errors.New("bbs watch failed")

				// avoid issues with race detector when the next test's
				// BeforeEach resets the changes channel
				changeChan := actualLRPCreates
				actualLRPToSend := actualLRP
				go func() { changeChan <- actualLRPToSend }()
			})

			It("should retry after the wait duration", func() {
				timeProvider.Increment(retryWaitDuration / 2)
				Consistently(hub.EmitCallCount).Should(BeZero())
				timeProvider.Increment(retryWaitDuration * 2)
				Eventually(hub.EmitCallCount).Should(Equal(1))
			})

			It("should be possible to SIGINT the route emitter", func() {
				process.Signal(os.Interrupt)
				Eventually(process.Wait()).Should(Receive())
			})
		})
	})
})
