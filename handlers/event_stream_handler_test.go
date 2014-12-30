package handlers_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/cloudfoundry-incubator/receptor/event"
	"github.com/cloudfoundry-incubator/receptor/event/eventfakes"
	"github.com/cloudfoundry-incubator/receptor/handlers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager"
	"github.com/vito/go-sse/sse"
)

type fakeEvent struct {
	Token string `json:"token"`
}

func (fakeEvent) EventType() event.EventType {
	return "fake"
}

var _ = Describe("Event Stream Handlers", func() {
	var (
		logger  lager.Logger
		fakeHub *eventfakes.FakeHub

		handler *handlers.EventStreamHandler

		server *httptest.Server
	)

	BeforeEach(func() {
		fakeHub = new(eventfakes.FakeHub)

		logger = lager.NewLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))

		handler = handlers.NewEventStreamHandler(fakeHub, logger)
	})

	AfterEach(func() {
		if server != nil {
			server.Close()
		}
	})

	Describe("EventStream", func() {
		var (
			fakeSource *eventfakes.FakeEventSource

			eventsToEmit chan<- event.Event

			response *http.Response
		)

		BeforeEach(func() {
			events := make(chan event.Event, 10)
			eventsToEmit = events

			fakeSource = new(eventfakes.FakeEventSource)
			fakeSource.NextStub = func() (event.Event, error) {
				e, ok := <-events
				if !ok {
					return nil, errors.New("event stream ended")
				}

				return e, nil
			}

			fakeHub.SubscribeReturns(fakeSource)

			server = httptest.NewServer(http.HandlerFunc(handler.EventStream))
		})

		JustBeforeEach(func() {
			var err error
			response, err = http.Get(server.URL)
			Ω(err).ShouldNot(HaveOccurred())
		})

		AfterEach(func() {
			// don't care if it's already closed
			defer func() { recover() }()
			close(eventsToEmit)
		})

		It("emits events from the hub to the connection", func() {
			reader := sse.NewReader(response.Body)

			eventsToEmit <- fakeEvent{"A"}

			Ω(reader.Next()).Should(Equal(sse.Event{
				Name: "fake",
				Data: []byte(`{"token":"A"}`),
			}))

			eventsToEmit <- fakeEvent{"B"}

			Ω(reader.Next()).Should(Equal(sse.Event{
				Name: "fake",
				Data: []byte(`{"token":"B"}`),
			}))

			close(eventsToEmit)

			_, err := reader.Next()
			Ω(err).Should(Equal(io.EOF))
		})
	})
})
