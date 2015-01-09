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

		desiredLRPChanges chan models.DesiredLRPChange
		desiredLRPErrors  chan error

		actualLRPChanges chan models.ActualLRPChange
		actualLRPErrors  chan error
	)

	BeforeEach(func() {
		bbs = new(fake_bbs.FakeReceptorBBS)
		hub = new(eventfakes.FakeHub)
		timeProvider = faketimeprovider.New(time.Now())
		logger := lagertest.NewTestLogger("test")

		desiredLRPChanges = make(chan models.DesiredLRPChange)
		desiredLRPErrors = make(chan error)

		actualLRPChanges = make(chan models.ActualLRPChange)
		actualLRPErrors = make(chan error)

		bbs.WatchForDesiredLRPChangesReturns(desiredLRPChanges, nil, desiredLRPErrors)
		bbs.WatchForActualLRPChangesReturns(actualLRPChanges, nil, actualLRPErrors)

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

		Context("when a create/update (includes an after) change arrives", func() {
			BeforeEach(func() {
				desiredChange := models.DesiredLRPChange{
					Before: nil,
					After:  &desiredLRP,
				}

				desiredLRPChanges <- desiredChange
			})

			It("emits a DesiredLRPChangedEvent to the hub", func() {
				Eventually(hub.EmitCallCount).Should(Equal(1))
				event := hub.EmitArgsForCall(0)

				desiredLRPChangedEvent, ok := event.(receptor.DesiredLRPChangedEvent)
				Ω(ok).Should(BeTrue())
				Ω(desiredLRPChangedEvent.DesiredLRPResponse).Should(Equal(serialization.DesiredLRPToResponse(desiredLRP)))
			})
		})

		Context("when the change is a delete (no after)", func() {
			BeforeEach(func() {
				desiredChange := models.DesiredLRPChange{
					Before: &desiredLRP,
					After:  nil,
				}

				desiredLRPChanges <- desiredChange
			})

			It("emits a DesiredLRPRemovedEvent to the hub", func() {
				Eventually(hub.EmitCallCount).Should(Equal(1))
				event := hub.EmitArgsForCall(0)

				desiredLRPRemovedEvent, ok := event.(receptor.DesiredLRPRemovedEvent)
				Ω(ok).Should(BeTrue())
				Ω(desiredLRPRemovedEvent.ProcessGuid).Should(Equal(desiredLRP.ProcessGuid))
			})
		})

		Context("when watching for change fails", func() {
			BeforeEach(func() {
				desiredLRPErrors <- errors.New("bbs watch failed")

				desiredChange := models.DesiredLRPChange{
					Before: nil,
					After:  &desiredLRP,
				}

				// avoid issues with race detector when the next test's
				// BeforeEach resets the changes channel
				changeChan := desiredLRPChanges
				go func() { changeChan <- desiredChange }()
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

		Context("when a create/update (includes an after) change arrives", func() {
			BeforeEach(func() {
				actualChange := models.ActualLRPChange{
					Before: nil,
					After:  &actualLRP,
				}

				actualLRPChanges <- actualChange
			})

			It("emits an ActualLRPChangedEvent to the hub", func() {
				Eventually(hub.EmitCallCount).Should(Equal(1))
				event := hub.EmitArgsForCall(0)

				actualLRPChangedEvent, ok := event.(receptor.ActualLRPChangedEvent)
				Ω(ok).Should(BeTrue())
				Ω(actualLRPChangedEvent.ActualLRPResponse).Should(Equal(serialization.ActualLRPToResponse(actualLRP)))
			})
		})

		Context("when the change is a delete (no after)", func() {
			BeforeEach(func() {
				actualChange := models.ActualLRPChange{
					Before: &actualLRP,
					After:  nil,
				}

				actualLRPChanges <- actualChange
			})

			It("emits an ActualLRPRemovedEvent to the hub", func() {
				Eventually(hub.EmitCallCount).Should(Equal(1))
				event := hub.EmitArgsForCall(0)

				actualLRPRemovedEvent, ok := event.(receptor.ActualLRPRemovedEvent)
				Ω(ok).Should(BeTrue())
				Ω(actualLRPRemovedEvent.ProcessGuid).Should(Equal(actualLRP.ProcessGuid))
				Ω(actualLRPRemovedEvent.Index).Should(Equal(actualLRP.Index))
			})
		})

		Context("when watching for change fails", func() {
			BeforeEach(func() {
				actualLRPErrors <- errors.New("bbs watch failed")

				actualChange := models.ActualLRPChange{
					Before: nil,
					After:  &actualLRP,
				}

				// avoid issues with race detector when the next test's
				// BeforeEach resets the changes channel
				changeChan := actualLRPChanges
				go func() { changeChan <- actualChange }()
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
