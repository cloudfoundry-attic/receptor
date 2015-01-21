package event_test

import (
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/event"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type fakeEvent struct {
	Token int `json:"token"`
}

func (fakeEvent) EventType() receptor.EventType {
	return "fake"
}

var _ = Describe("Hub", func() {
	var (
		hub event.Hub
	)

	BeforeEach(func() {
		hub = event.NewHub()
	})

	Describe("HasSubscribers", func() {
		It("returns false", func() {
			Ω(hub.HasSubscribers()).Should(BeFalse())
		})

		Context("when there is a subscriber", func() {
			var eventSource receptor.EventSource

			BeforeEach(func() {
				var err error
				eventSource, err = hub.Subscribe()
				Ω(err).ShouldNot(HaveOccurred())
			})

			It("returns true", func() {
				Ω(hub.HasSubscribers()).Should(BeTrue())
			})

			Context("when all subscribers are dropped", func() {
				BeforeEach(func() {
					err := eventSource.Close()
					Ω(err).ShouldNot(HaveOccurred())

					// emit event so hub sees closed source and drops subscription
					hub.Emit(fakeEvent{})
				})

				It("returns false", func() {
					Ω(hub.HasSubscribers()).Should(BeFalse())
				})
			})
		})
	})

	It("fans-out events emitted to it to all subscribers", func() {
		source1, err := hub.Subscribe()
		Ω(err).ShouldNot(HaveOccurred())
		source2, err := hub.Subscribe()
		Ω(err).ShouldNot(HaveOccurred())

		hub.Emit(fakeEvent{Token: 1})
		Ω(source1.Next()).Should(Equal(fakeEvent{Token: 1}))
		Ω(source2.Next()).Should(Equal(fakeEvent{Token: 1}))

		hub.Emit(fakeEvent{Token: 2})
		Ω(source1.Next()).Should(Equal(fakeEvent{Token: 2}))
		Ω(source2.Next()).Should(Equal(fakeEvent{Token: 2}))
	})

	It("closes slow consumers after MAX_PENDING_SUBSCRIBER_EVENTS missed events", func() {
		slowConsumer, err := hub.Subscribe()
		Ω(err).ShouldNot(HaveOccurred())

		By("filling the 'buffer'")
		for eventToken := 0; eventToken < event.MAX_PENDING_SUBSCRIBER_EVENTS; eventToken++ {
			hub.Emit(fakeEvent{Token: eventToken})
		}

		By("reading 2 events off")
		ev, err := slowConsumer.Next()
		Ω(err).ShouldNot(HaveOccurred())
		Ω(ev).Should(Equal(fakeEvent{Token: 0}))

		ev, err = slowConsumer.Next()
		Ω(err).ShouldNot(HaveOccurred())
		Ω(ev).Should(Equal(fakeEvent{Token: 1}))

		By("putting 3 more events on, 'overflowing the buffer' and making the consumer 'slow'")
		for eventToken := event.MAX_PENDING_SUBSCRIBER_EVENTS; eventToken < event.MAX_PENDING_SUBSCRIBER_EVENTS+3; eventToken++ {
			hub.Emit(fakeEvent{Token: eventToken})
		}

		By("reading off all the 'buffered' events")
		for eventToken := 2; eventToken < event.MAX_PENDING_SUBSCRIBER_EVENTS+2; eventToken++ {
			ev, err = slowConsumer.Next()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(ev).Should(Equal(fakeEvent{Token: eventToken}))
		}

		By("trying to read more out of the source")
		_, err = slowConsumer.Next()
		Ω(err).Should(Equal(receptor.ErrReadFromClosedSource))
	})

	Describe("closing an event source", func() {
		It("prevents current events from propagating to the source", func() {
			source, err := hub.Subscribe()
			Ω(err).ShouldNot(HaveOccurred())

			hub.Emit(fakeEvent{Token: 1})
			Ω(source.Next()).Should(Equal(fakeEvent{Token: 1}))

			err = source.Close()
			Ω(err).ShouldNot(HaveOccurred())

			_, err = source.Next()
			Ω(err).Should(Equal(receptor.ErrReadFromClosedSource))
		})

		It("prevents future events from propagating to the source", func() {
			source, err := hub.Subscribe()
			Ω(err).ShouldNot(HaveOccurred())

			err = source.Close()
			Ω(err).ShouldNot(HaveOccurred())

			hub.Emit(fakeEvent{Token: 1})

			_, err = source.Next()
			Ω(err).Should(Equal(receptor.ErrReadFromClosedSource))
		})

		Context("when the source is already closed", func() {
			It("errors", func() {
				source, err := hub.Subscribe()
				Ω(err).ShouldNot(HaveOccurred())

				err = source.Close()
				Ω(err).ShouldNot(HaveOccurred())

				err = source.Close()
				Ω(err).Should(Equal(receptor.ErrSourceAlreadyClosed))
			})
		})
	})

	Describe("closing the hub", func() {
		It("all subscribers receive errors", func() {
			source, err := hub.Subscribe()
			Ω(err).ShouldNot(HaveOccurred())

			err = hub.Close()
			Ω(err).ShouldNot(HaveOccurred())

			_, err = source.Next()
			Ω(err).Should(Equal(receptor.ErrReadFromClosedSource))
		})

		It("does not accept new subscribers", func() {
			err := hub.Close()
			Ω(err).ShouldNot(HaveOccurred())

			_, err = hub.Subscribe()
			Ω(err).Should(Equal(receptor.ErrSubscribedToClosedHub))
		})

		Context("when the hub is already closed", func() {
			BeforeEach(func() {
				err := hub.Close()
				Ω(err).ShouldNot(HaveOccurred())
			})

			It("errors", func() {
				err := hub.Close()
				Ω(err).Should(Equal(receptor.ErrHubAlreadyClosed))
			})
		})
	})
})
