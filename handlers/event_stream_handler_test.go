package handlers_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/event/eventfakes"
	"github.com/cloudfoundry-incubator/receptor/fake_receptor"
	"github.com/cloudfoundry-incubator/receptor/handlers"
	"github.com/pivotal-golang/lager"
	"github.com/vito/go-sse/sse"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type fakeEvent struct {
	Token string `json:"token"`
}

func (fakeEvent) EventType() receptor.EventType {
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
			fakeSource *fake_receptor.FakeEventSource

			eventsToEmit chan<- receptor.Event

			response *http.Response
		)

		BeforeEach(func() {
			events := make(chan receptor.Event, 10)
			eventsToEmit = events

			fakeSource = new(fake_receptor.FakeEventSource)
			fakeSource.NextStub = func() (receptor.Event, error) {
				e, ok := <-events
				if !ok {
					return nil, errors.New("event stream ended")
				}

				return e, nil
			}

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

		Context("when failing to subscribe to the event hub", func() {
			BeforeEach(func() {
				fakeHub.SubscribeReturns(nil, errors.New("failed-to-subscribe"))
			})

			It("returns an internal server error", func() {
				Ω(response.StatusCode).Should(Equal(http.StatusInternalServerError))
			})
		})

		Context("when successfully subscribing to the event hub", func() {
			BeforeEach(func() {
				fakeHub.SubscribeReturns(fakeSource, nil)
			})

			It("emits events from the hub to the connection", func() {
				reader := sse.NewReader(response.Body)

				eventsToEmit <- fakeEvent{"A"}

				Ω(reader.Next()).Should(Equal(sse.Event{
					ID:   "0",
					Name: "fake",
					Data: []byte(`{"token":"A"}`),
				}))

				eventsToEmit <- fakeEvent{"B"}

				Ω(reader.Next()).Should(Equal(sse.Event{
					ID:   "1",
					Name: "fake",
					Data: []byte(`{"token":"B"}`),
				}))

				close(eventsToEmit)

				_, err := reader.Next()
				Ω(err).Should(Equal(io.EOF))
			})
		})
	})
})
