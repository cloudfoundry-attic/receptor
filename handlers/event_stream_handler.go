package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/cloudfoundry-incubator/receptor/event"
	"github.com/pivotal-golang/lager"
	"github.com/vito/go-sse/sse"
)

type EventStreamHandler struct {
	hub    event.Hub
	logger lager.Logger
}

func NewEventStreamHandler(hub event.Hub, logger lager.Logger) *EventStreamHandler {
	return &EventStreamHandler{
		hub:    hub,
		logger: logger.Session("event-stream-handler"),
	}
}

func (h *EventStreamHandler) EventStream(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("sink")

	flusher := w.(http.Flusher)

	source := h.hub.Subscribe()

	w.WriteHeader(http.StatusOK)

	flusher.Flush()

	for {
		e, err := source.Next()
		if err != nil {
			logger.Error("failed-to-get-next-event", err)
			return
		}

		payload, err := json.Marshal(e)
		if err != nil {
			logger.Error("failed-to-marshal-event", err)
			return
		}

		err = sse.Event{
			Name: string(e.EventType()),
			Data: payload,
		}.Write(w)
		if err != nil {
			break
		}

		flusher.Flush()
	}
}
