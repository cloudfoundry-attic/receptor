package event

import (
	"errors"
	"sync"
)

const MAX_PENDING_SUBSCRIBER_EVENTS = 1024

var ErrSlowConsumer = errors.New("slow consumer")
var ErrReadFromClosedSource = errors.New("read from closed source")

//go:generate counterfeiter -o eventfakes/fake_hub.go . Hub
type Hub interface {
	Emit(Event)
	Subscribe() EventSource
}

type hub struct {
	subscribers  map[chan Event]*hubSource
	subscribersL sync.Mutex
}

func NewHub() Hub {
	return &hub{
		subscribers: make(map[chan Event]*hubSource),
	}
}

func (hub *hub) Subscribe() EventSource {
	ch := make(chan Event, MAX_PENDING_SUBSCRIBER_EVENTS)

	source := &hubSource{
		events: ch,

		unsubscribe: func() {
			hub.subscribersL.Lock()
			delete(hub.subscribers, ch)
			hub.subscribersL.Unlock()

			close(ch)
		},
	}

	hub.subscribersL.Lock()
	hub.subscribers[ch] = source
	hub.subscribersL.Unlock()

	return source
}

func (hub *hub) Emit(event Event) {
	hub.subscribersL.Lock()

	for sub, source := range hub.subscribers {
		select {
		case sub <- event:
		default:
			source.err = ErrSlowConsumer
			close(sub)
		}
	}

	hub.subscribersL.Unlock()
}

type hubSource struct {
	events <-chan Event
	err    error

	unsubscribe func()

	lock sync.Mutex
}

func (source *hubSource) Next() (Event, error) {
	e, ok := <-source.events
	if !ok {
		return nil, source.err
	}

	return e, nil
}

func (source *hubSource) Close() {
	source.err = ErrReadFromClosedSource
	source.unsubscribe()
}
