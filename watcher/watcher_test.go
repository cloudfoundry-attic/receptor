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

		desiredLRPStop   chan bool
		desiredLRPErrors chan error

		actualLRPStop   chan bool
		actualLRPErrors chan error
	)

	BeforeEach(func() {
		bbs = new(fake_bbs.FakeReceptorBBS)
		hub = new(eventfakes.FakeHub)
		timeProvider = faketimeprovider.New(time.Now())
		logger := lagertest.NewTestLogger("test")

		desiredLRPStop = make(chan bool, 1)
		desiredLRPErrors = make(chan error)

		actualLRPStop = make(chan bool, 1)
		actualLRPErrors = make(chan error)

		bbs.WatchForDesiredLRPChangesReturns(desiredLRPStop, desiredLRPErrors)
		bbs.WatchForActualLRPChangesReturns(actualLRPStop, actualLRPErrors)

		receptorWatcher = watcher.NewWatcher(bbs, hub, timeProvider, retryWaitDuration, logger)
	})

	AfterEach(func() {
		process.Signal(os.Interrupt)
		Eventually(process.Wait()).Should(Receive())
	})

	Describe("starting", func() {
		Context("when the hub initially reports no subscribers", func() {
			BeforeEach(func() {
				hub.RegisterCallbackStub = func(cb func(int)) {
					cb(0)
				}
				process = ifrit.Invoke(receptorWatcher)
			})

			It("does not request a watch", func() {
				Consistently(bbs.WatchForDesiredLRPChangesCallCount).Should(BeZero())
				Consistently(bbs.WatchForActualLRPChangesCallCount).Should(BeZero())
			})

			Context("and then the hub reports a subscriber", func() {
				var callback func(int)

				BeforeEach(func() {
					Ω(hub.RegisterCallbackCallCount()).Should(Equal(1))
					callback = hub.RegisterCallbackArgsForCall(0)
					callback(1)
				})

				It("requests watches", func() {
					Eventually(bbs.WatchForDesiredLRPChangesCallCount).Should(Equal(1))
					Eventually(bbs.WatchForActualLRPChangesCallCount).Should(Equal(1))
				})

				Context("and then the hub reports two subscribers", func() {
					BeforeEach(func() {
						callback(2)
					})

					It("does not request more watches", func() {
						Eventually(bbs.WatchForDesiredLRPChangesCallCount).Should(Equal(1))
						Consistently(bbs.WatchForDesiredLRPChangesCallCount).Should(Equal(1))

						Eventually(bbs.WatchForActualLRPChangesCallCount).Should(Equal(1))
						Consistently(bbs.WatchForActualLRPChangesCallCount).Should(Equal(1))
					})
				})

				Context("and then the hub reports no subscribers", func() {
					BeforeEach(func() {
						callback(0)
					})

					It("stops the watches", func() {
						Eventually(desiredLRPStop).Should(Receive())
						Eventually(actualLRPStop).Should(Receive())
					})
				})

				Context("when the desired watch reports an error", func() {
					BeforeEach(func() {
						desiredLRPErrors <- errors.New("oh no!")
					})

					It("requests a new desired watch after the retry interval", func() {
						timeProvider.Increment(retryWaitDuration / 2)
						Consistently(bbs.WatchForDesiredLRPChangesCallCount).Should(Equal(1))
						timeProvider.Increment(retryWaitDuration * 2)
						Eventually(bbs.WatchForDesiredLRPChangesCallCount).Should(Equal(2))
					})

					Context("and the hub reports no subscribers before the retry interval elapses", func() {
						BeforeEach(func() {
							timeProvider.Increment(retryWaitDuration / 2)
							callback(0)
							// give watcher time to clear out event loop
							time.Sleep(10 * time.Millisecond)
						})

						It("does not request new watches", func() {
							timeProvider.Increment(retryWaitDuration * 2)
							Consistently(bbs.WatchForDesiredLRPChangesCallCount).Should(Equal(1))
						})
					})
				})

				Context("when the actual watch reports an error", func() {
					BeforeEach(func() {
						actualLRPErrors <- errors.New("oh no!")
					})

					It("requests a new actual watch after the retry interval", func() {
						timeProvider.Increment(retryWaitDuration / 2)
						Consistently(bbs.WatchForActualLRPChangesCallCount).Should(Equal(1))
						timeProvider.Increment(retryWaitDuration * 2)
						Eventually(bbs.WatchForActualLRPChangesCallCount).Should(Equal(2))
					})

					Context("and the hub reports no subscribers before the retry interval elapses", func() {
						BeforeEach(func() {
							timeProvider.Increment(retryWaitDuration / 2)
							callback(0)
							// give watcher time to clear out event loop
							time.Sleep(10 * time.Millisecond)
						})

						It("does not request new watches", func() {
							timeProvider.Increment(retryWaitDuration * 2)
							Consistently(bbs.WatchForActualLRPChangesCallCount).Should(Equal(1))
						})
					})
				})
			})
		})

		Context("when the hub initially reports a subscriber", func() {
			BeforeEach(func() {
				hub.RegisterCallbackStub = func(cb func(int)) {
					cb(1)
				}
				process = ifrit.Invoke(receptorWatcher)
			})

			It("requests watches", func() {
				Eventually(bbs.WatchForDesiredLRPChangesCallCount).Should(Equal(1))
				Eventually(bbs.WatchForActualLRPChangesCallCount).Should(Equal(1))
			})

			Context("and then the watcher is signaled to stop", func() {
				It("stops the watches", func() {
					process.Signal(os.Interrupt)
					Eventually(desiredLRPStop).Should(Receive())
					Eventually(actualLRPStop).Should(Receive())
					Eventually(process.Wait()).Should(Receive())
				})
			})
		})
	})

	Describe("when watching the bbs", func() {
		var (
			desiredCreateCB func(models.DesiredLRP)
			desiredChangeCB func(models.DesiredLRPChange)
			desiredDeleteCB func(models.DesiredLRP)
			actualCreateCB  func(models.ActualLRP)
			actualChangeCB  func(models.ActualLRPChange)
			actualDeleteCB  func(models.ActualLRP)
		)

		BeforeEach(func() {
			hub.RegisterCallbackStub = func(cb func(int)) {
				cb(1)
			}
			process = ifrit.Invoke(receptorWatcher)
			Eventually(bbs.WatchForDesiredLRPChangesCallCount).Should(Equal(1))
			Eventually(bbs.WatchForActualLRPChangesCallCount).Should(Equal(1))

			_, desiredCreateCB, desiredChangeCB, desiredDeleteCB = bbs.WatchForDesiredLRPChangesArgsForCall(0)
			_, actualCreateCB, actualChangeCB, actualDeleteCB = bbs.WatchForActualLRPChangesArgsForCall(0)
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
					desiredCreateCB(desiredLRP)
				})

				It("emits a DesiredLRPCreatedEvent to the hub", func() {
					Ω(hub.EmitCallCount()).Should(Equal(1))
					event := hub.EmitArgsForCall(0)

					desiredLRPCreatedEvent, ok := event.(receptor.DesiredLRPCreatedEvent)
					Ω(ok).Should(BeTrue())
					Ω(desiredLRPCreatedEvent.DesiredLRPResponse).Should(Equal(serialization.DesiredLRPToResponse(desiredLRP)))
				})
			})

			Context("when a change arrives", func() {
				BeforeEach(func() {
					desiredChangeCB(models.DesiredLRPChange{Before: desiredLRP, After: desiredLRP})
				})

				It("emits a DesiredLRPChangedEvent to the hub", func() {
					Ω(hub.EmitCallCount()).Should(Equal(1))
					event := hub.EmitArgsForCall(0)

					desiredLRPChangedEvent, ok := event.(receptor.DesiredLRPChangedEvent)
					Ω(ok).Should(BeTrue())
					Ω(desiredLRPChangedEvent.Before).Should(Equal(serialization.DesiredLRPToResponse(desiredLRP)))
					Ω(desiredLRPChangedEvent.After).Should(Equal(serialization.DesiredLRPToResponse(desiredLRP)))
				})
			})

			Context("when a delete arrives", func() {
				BeforeEach(func() {
					desiredDeleteCB(desiredLRP)
				})

				It("emits a DesiredLRPRemovedEvent to the hub", func() {
					Ω(hub.EmitCallCount()).Should(Equal(1))
					event := hub.EmitArgsForCall(0)

					desiredLRPRemovedEvent, ok := event.(receptor.DesiredLRPRemovedEvent)
					Ω(ok).Should(BeTrue())
					Ω(desiredLRPRemovedEvent.DesiredLRPResponse).Should(Equal(serialization.DesiredLRPToResponse(desiredLRP)))
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
					actualCreateCB(actualLRP)
				})

				It("emits an ActualLRPCreatedEvent to the hub", func() {
					Ω(hub.EmitCallCount()).Should(Equal(1))
					event := hub.EmitArgsForCall(0)

					actualLRPCreatedEvent, ok := event.(receptor.ActualLRPCreatedEvent)
					Ω(ok).Should(BeTrue())
					Ω(actualLRPCreatedEvent.ActualLRPResponse).Should(Equal(serialization.ActualLRPToResponse(actualLRP)))
				})
			})

			Context("when a change arrives", func() {
				BeforeEach(func() {
					actualChangeCB(models.ActualLRPChange{Before: actualLRP, After: actualLRP})
				})

				It("emits an ActualLRPChangedEvent to the hub", func() {
					Ω(hub.EmitCallCount()).Should(Equal(1))
					event := hub.EmitArgsForCall(0)

					actualLRPChangedEvent, ok := event.(receptor.ActualLRPChangedEvent)
					Ω(ok).Should(BeTrue())
					Ω(actualLRPChangedEvent.Before).Should(Equal(serialization.ActualLRPToResponse(actualLRP)))
					Ω(actualLRPChangedEvent.After).Should(Equal(serialization.ActualLRPToResponse(actualLRP)))
				})
			})

			Context("when a delete arrives", func() {
				BeforeEach(func() {
					actualDeleteCB(actualLRP)
				})

				It("emits an ActualLRPRemovedEvent to the hub", func() {
					Ω(hub.EmitCallCount()).Should(Equal(1))
					event := hub.EmitArgsForCall(0)

					actualLRPRemovedEvent, ok := event.(receptor.ActualLRPRemovedEvent)
					Ω(ok).Should(BeTrue())
					Ω(actualLRPRemovedEvent.ActualLRPResponse).Should(Equal(serialization.ActualLRPToResponse(actualLRP)))
				})
			})
		})
	})
})
