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

	RegisterCallback(func(count int))
}

type hub struct {
	subscribers []*hubSource
	closed      bool
	lock        sync.Mutex

	cb func(count int)
}

func NewHub() Hub {
	return &hub{}
}

func (hub *hub) RegisterCallback(cb func(int)) {
	hub.lock.Lock()
	hub.cb = cb
	size := len(hub.subscribers)
	hub.lock.Unlock()
	if cb != nil {
		cb(size)
	}
}

func (hub *hub) Subscribe() (receptor.EventSource, error) {
	hub.lock.Lock()

	if hub.closed {
		hub.lock.Unlock()
		return nil, receptor.ErrSubscribedToClosedHub
	}

	sub := newSource(MAX_PENDING_SUBSCRIBER_EVENTS)
	hub.subscribers = append(hub.subscribers, sub)
	cb := hub.cb
	size := len(hub.subscribers)
	hub.lock.Unlock()

	if cb != nil {
		cb(size)
	}
	return sub, nil
}

func (hub *hub) Emit(event receptor.Event) {
	hub.lock.Lock()

	size := len(hub.subscribers)
	remainingSubscribers := make([]*hubSource, 0, len(hub.subscribers))

	for _, sub := range hub.subscribers {
		err := sub.send(event)
		if err == nil {
			remainingSubscribers = append(remainingSubscribers, sub)
		}
	}

	hub.subscribers = remainingSubscribers
	var cb func(int)
	if len(hub.subscribers) != size {
		cb = hub.cb
		size = len(hub.subscribers)
	}
	hub.lock.Unlock()

	if cb != nil {
		cb(size)
	}
}

func (hub *hub) Close() error {
	hub.lock.Lock()
	defer hub.lock.Unlock()

	if hub.closed {
		return receptor.ErrHubAlreadyClosed
	}

	hub.closeSubscribers()
	hub.closed = true
	if hub.cb != nil {
		hub.cb(0)
	}
	return nil
}

func (hub *hub) closeSubscribers() {
	for _, sub := range hub.subscribers {
		_ = sub.Close()
	}
	hub.subscribers = nil
}

type hubSource struct {
	events chan receptor.Event
	closed bool
	lock   sync.Mutex
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
	source.lock.Lock()
	defer source.lock.Unlock()

	if source.closed {
		return receptor.ErrSourceAlreadyClosed
	}
	close(source.events)
	source.closed = true
	return nil
}

func (source *hubSource) send(event receptor.Event) error {
	source.lock.Lock()

	if source.closed {
		source.lock.Unlock()
		return receptor.ErrSendToClosedSource
	}

	select {
	case source.events <- event:
		source.lock.Unlock()
		return nil

	default:
		source.lock.Unlock()
		err := source.Close()
		if err != nil {
			return err
		}

		return receptor.ErrSlowConsumer
	}
}
