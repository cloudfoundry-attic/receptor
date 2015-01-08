package event

import (
	"sync"

	"github.com/cloudfoundry-incubator/receptor"
)

const MAX_PENDING_SUBSCRIBER_EVENTS = 1024

//go:generate counterfeiter -o eventfakes/fake_hub.go . Hub
type Hub interface {
	Emit(receptor.Event)
	Subscribe() receptor.EventSource
}

type hub struct {
	subscribers  map[chan receptor.Event]*hubSource
	subscribersL sync.Mutex
}

func NewHub() Hub {
	return &hub{
		subscribers: make(map[chan receptor.Event]*hubSource),
	}
}

func (hub *hub) Subscribe() receptor.EventSource {
	ch := make(chan receptor.Event, MAX_PENDING_SUBSCRIBER_EVENTS)

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

func (hub *hub) Emit(event receptor.Event) {
	hub.subscribersL.Lock()

	for sub, source := range hub.subscribers {
		select {
		case sub <- event:
		default:
			source.err = receptor.ErrSlowConsumer
			close(sub)
		}
	}

	hub.subscribersL.Unlock()
}

type hubSource struct {
	events <-chan receptor.Event
	err    error

	unsubscribe func()

	lock sync.Mutex
}

func (source *hubSource) Next() (receptor.Event, error) {
	e, ok := <-source.events
	if !ok {
		return nil, source.err
	}

	return e, nil
}

func (source *hubSource) Close() {
	source.err = receptor.ErrReadFromClosedSource
	source.unsubscribe()
}
