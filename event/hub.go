package event

type Hub interface {
	Subscribe() (EventSource, error)
}

type EventSource interface {
	Next() (Event, error)
	Close() error
}

type Event interface {
	EventType() EventType
}

type EventType string
