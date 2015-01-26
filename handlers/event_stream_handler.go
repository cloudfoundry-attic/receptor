package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/cloudfoundry-incubator/receptor"
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
	closeNotifier := w.(http.CloseNotifier).CloseNotify()

	flusher := w.(http.Flusher)

	source, err := h.hub.Subscribe()
	if err != nil {
		logger.Error("failed-to-subscribe-to-event-hub", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() {
		err := source.Close()
		if err != nil {
			logger.Debug("failed-to-close-event-source", lager.Data{"error-msg": err.Error()})
		}
	}()

	w.WriteHeader(http.StatusOK)

	flusher.Flush()

	eventID := 0
	errChan := make(chan error, 1)
	eventChan := make(chan receptor.Event)

	for {
		go func() {
			e, err := source.Next()
			if err != nil {
				errChan <- err
			} else if e != nil {
				eventChan <- e
			}
		}()

		select {
		case event := <-eventChan:
			payload, err := json.Marshal(event)
			if err != nil {
				logger.Error("failed-to-marshal-event", err)
				return
			}

			err = sse.Event{
				ID:   strconv.Itoa(eventID),
				Name: string(event.EventType()),
				Data: payload,
			}.Write(w)
			if err != nil {
				break
			}

			flusher.Flush()

			eventID++

		case err := <-errChan:
			logger.Error("failed-to-get-next-event", err)
			return

		case <-closeNotifier:
			logger.Info("client-closed-response-body")
			return
		}
	}
}
