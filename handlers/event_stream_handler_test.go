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

type unmarshalableEvent struct {
	Fn func() `json:"fn"`
}

func (unmarshalableEvent) EventType() receptor.EventType {
	return "unmarshalable"
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

			response        *http.Response
			eventStreamDone chan struct{}
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

			eventStreamDone = make(chan struct{})
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handler.EventStream(w, r)
				close(eventStreamDone)
			}))
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
				reader := sse.NewReadCloser(response.Body)

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
			})

			Context("when the source provides an unmarshalable event", func() {
				BeforeEach(func() {
					unmarshalable := unmarshalableEvent{Fn: func() {}}
					eventsToEmit <- unmarshalable
				})

				It("closes the event stream to the client", func() {
					reader := sse.NewReadCloser(response.Body)
					_, err := reader.Next()
					Ω(err).Should(Equal(io.EOF))
				})

				It("closes the event source", func() {
					Eventually(fakeSource.CloseCallCount).Should(Equal(1))
				})
			})

			Context("when the event source returns an error", func() {
				BeforeEach(func() {
					close(eventsToEmit)
				})

				It("closes the client event stream", func() {
					reader := sse.NewReadCloser(response.Body)
					_, err := reader.Next()
					Ω(err).Should(Equal(io.EOF))
				})

				It("close the event source", func() {
					Eventually(fakeSource.CloseCallCount).Should(Equal(1))
				})
			})

			Context("when the client closes the response body", func() {
				It("closes its connection to the hub", func() {
					reader := sse.NewReadCloser(response.Body)
					eventsToEmit <- fakeEvent{"A"}
					err := reader.Close()
					Ω(err).ShouldNot(HaveOccurred())
					Eventually(fakeSource.CloseCallCount).Should(Equal(1))
				})

				It("returns early", func() {
					reader := sse.NewReadCloser(response.Body)
					eventsToEmit <- fakeEvent{"A"}
					err := reader.Close()
					Ω(err).ShouldNot(HaveOccurred())
					Eventually(eventStreamDone).Should(BeClosed())
				})
			})
		})
	})
})
