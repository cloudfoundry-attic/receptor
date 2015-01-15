package event

import (
	"sync"

	"github.com/cloudfoundry-incubator/receptor"
)

const MAX_PENDING_SUBSCRIBER_EVENTS = 1024

//go:generate counterfeiter -o eventfakes/fake_hub.go . Hub
type Hub interface {
	Subscribe() (receptor.EventSource, error)
	Emit(receptor.Event)
	Close() error
}

type hub struct {
	subscribers []*hubSource
	closed      bool
	lock        sync.Mutex
}

func NewHub() Hub {
	return &hub{}
}

func (hub *hub) Subscribe() (receptor.EventSource, error) {
	hub.lock.Lock()
	defer hub.lock.Unlock()

	if hub.closed {
		return nil, receptor.ErrSubscribedToClosedHub
	}

	sub := newSource(MAX_PENDING_SUBSCRIBER_EVENTS)
	hub.subscribers = append(hub.subscribers, sub)
	return sub, nil
}

func (hub *hub) Emit(event receptor.Event) {
	hub.lock.Lock()
	defer hub.lock.Unlock()

	remainingSubscribers := make([]*hubSource, 0, len(hub.subscribers))

	for _, sub := range hub.subscribers {
		err := sub.send(event)
		if err == nil {
			remainingSubscribers = append(remainingSubscribers, sub)
		}
	}

	hub.subscribers = remainingSubscribers
}

func (hub *hub) Close() error {
	hub.lock.Lock()
	defer hub.lock.Unlock()

	if hub.closed {
		return receptor.ErrHubAlreadyClosed
	}

	hub.closeSubscribers()
	hub.closed = true
	return nil
}

func (hub *hub) closeSubscribers() {
	for _, sub := range hub.subscribers {
		_ = sub.Close()
	}
	hub.subscribers = nil
}

type hubSource struct {
	events    chan receptor.Event
	closed    bool
	closeLock sync.Mutex
}

func newSource(maxPendingEvents int) *hubSource {
	return &hubSource{
		events: make(chan receptor.Event, maxPendingEvents),
	}
}

func (source *hubSource) Next() (receptor.Event, error) {
	event, ok := <-source.events
	if !ok {
		return nil, receptor.ErrReadFromClosedSource
	}
	return event, nil
}

func (source *hubSource) Close() error {
	source.closeLock.Lock()
	defer source.closeLock.Unlock()

	if source.closed {
		return receptor.ErrSourceAlreadyClosed
	}
	close(source.events)
	source.closed = true
	return nil
}

func (source *hubSource) send(event receptor.Event) error {
	select {
	case source.events <- event:
		return nil

	default:
		err := source.Close()
		if err != nil {
			return err
		}

		return receptor.ErrSlowConsumer
	}
}
