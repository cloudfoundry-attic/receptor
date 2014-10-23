package heartbeat

import (
	"encoding/json"
	"math"
	"time"
)

const (
	RouterGreetTopic      = "router.greet"
	RouterStartTopic      = "router.start"
	RouterRegisterTopic   = "router.register"
	RouterUnregisterTopic = "router.unregister"
)

type RegistryMessage struct {
	URIs []string `json:"uris"`
	Host string   `json:"host"`
	Port int      `json:"port"`
}

type GreetingMessage struct {
	Id               string
	Hosts            []string
	RegisterInterval time.Duration
}

// jsonGreetingMessage is used internally to marshal and unmarshal json
type jsonGreetingMessage struct {
	Id                               string   `json:"id"`
	Hosts                            []string `json:"hosts"`
	MinimumRegisterIntervalInSeconds int64    `json:"minimumRegisterIntervalInSeconds"`
}

func (msg GreetingMessage) MarshalJSON() ([]byte, error) {
	seconds := math.Ceil(msg.RegisterInterval.Seconds())

	jsonMsg := jsonGreetingMessage{
		Id:    msg.Id,
		Hosts: msg.Hosts,
		MinimumRegisterIntervalInSeconds: int64(seconds),
	}

	return json.Marshal(jsonMsg)
}

func (msg *GreetingMessage) UnmarshalJSON(data []byte) error {
	jsonMsg := jsonGreetingMessage{}
	err := json.Unmarshal(data, &jsonMsg)
	if err != nil {
		return err
	}

	msg.Id = jsonMsg.Id
	msg.Hosts = jsonMsg.Hosts
	msg.RegisterInterval = time.Duration(jsonMsg.MinimumRegisterIntervalInSeconds) * time.Second

	return nil
}
