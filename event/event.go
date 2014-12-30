package event

//go:generate counterfeiter -o eventfakes/fake_event_source.go . EventSource
type EventSource interface {
	Next() (Event, error)
	Close()
}

type Event interface {
	EventType() EventType
}

type EventType string
