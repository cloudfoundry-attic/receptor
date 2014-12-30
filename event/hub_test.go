package event_test

import (
	. "github.com/cloudfoundry-incubator/receptor/event"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type fakeEvent struct {
	Token int `json:"token"`
}

func (fakeEvent) EventType() EventType {
	return "fake"
}

var _ = Describe("Hub", func() {
	var (
		hub Hub
	)

	BeforeEach(func() {
		hub = NewHub()
	})

	It("fans-out events emitted to it to all subscribers", func() {
		source1 := hub.Subscribe()
		source2 := hub.Subscribe()

		hub.Emit(fakeEvent{Token: 1})
		Ω(source1.Next()).Should(Equal(fakeEvent{Token: 1}))
		Ω(source2.Next()).Should(Equal(fakeEvent{Token: 1}))

		hub.Emit(fakeEvent{Token: 2})
		Ω(source1.Next()).Should(Equal(fakeEvent{Token: 2}))
		Ω(source2.Next()).Should(Equal(fakeEvent{Token: 2}))
	})

	It("drops slow consumers after 1024 missed events", func() {
		slowConsumer := hub.Subscribe()
		nonSlowConsumer := hub.Subscribe()

		for i := 0; i < 1024; i++ {
			hub.Emit(fakeEvent{Token: i})
			Ω(nonSlowConsumer.Next()).Should(Equal(fakeEvent{Token: i}))
		}

		hub.Emit(fakeEvent{Token: 1024})

		for i := 0; i < 1024; i++ {
			Ω(slowConsumer.Next()).Should(Equal(fakeEvent{Token: i}))
		}

		_, err := slowConsumer.Next()
		Ω(err).Should(Equal(ErrSlowConsumer))

		Ω(nonSlowConsumer.Next()).Should(Equal(fakeEvent{Token: 1024}))
	})

	Describe("closing an event source", func() {
		It("prevents future events from propagating to the source", func() {
			source := hub.Subscribe()

			hub.Emit(fakeEvent{Token: 1})
			Ω(source.Next()).Should(Equal(fakeEvent{Token: 1}))

			source.Close()

			_, err := source.Next()
			Ω(err).Should(Equal(ErrReadFromClosedSource))
		})
	})
})
