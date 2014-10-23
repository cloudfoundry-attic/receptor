package heartbeat

import (
	"encoding/json"
	"os"
	"time"

	"github.com/apcera/nats"
	"github.com/cloudfoundry/gunk/diegonats"
	"github.com/pivotal-golang/lager"
)

const NatsGreetRequestTimeout = 5 * time.Second
const MaxConcurrentGreetings = 5

type Heartbeater struct {
	natsClient      diegonats.NATSClient
	registryMessage RegistryMessage
	initialInterval time.Duration
	logger          lager.Logger
}

func New(natsClient diegonats.NATSClient, registryMessage RegistryMessage, initialInterval time.Duration, logger lager.Logger) *Heartbeater {
	return &Heartbeater{
		natsClient:      natsClient,
		registryMessage: registryMessage,
		initialInterval: initialInterval,
		logger:          logger.Session("nats-heartbeat", lager.Data{"registration": registryMessage}),
	}
}

func (h Heartbeater) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	h.logger.Info("starting")

	registerPayload, err := json.Marshal(h.registryMessage)
	if err != nil {
		return err
	}

	err = h.natsClient.Publish(RouterRegisterTopic, registerPayload)
	h.logger.Debug("initial-registration")
	if err != nil {
		return err
	}

	greetChan := make(greetingChannel, MaxConcurrentGreetings)

	routerSubscription, err := h.natsClient.Subscribe(RouterStartTopic, greetChan.SendGreeting)
	if err != nil {
		return err
	}

	go greetChan.GreetRouter(h.natsClient)

	heartbeatTimer := time.NewTicker(h.initialInterval)

	close(ready)

	for {
		select {
		case greeting := <-greetChan:
			heartbeatTimer = time.NewTicker(greeting.RegisterInterval)

		case <-heartbeatTimer.C:
			err := h.natsClient.Publish(RouterRegisterTopic, registerPayload)
			if err != nil {
				h.logger.Error("publish-registration-failed", err)
			} else {
				h.logger.Debug("published-registration")
			}

		case sig := <-signals:
			if sig == os.Kill {
				h.logger.Info("killed")
				return nil
			}
			h.natsClient.Unsubscribe(routerSubscription)
			err := h.natsClient.Publish(RouterUnregisterTopic, registerPayload)
			if err != nil {
				h.logger.Error("unregistering-failed", err)
			} else {
				h.logger.Info("unregistered")
			}
			return nil
		}
	}
}

type greetingChannel chan *GreetingMessage

func (c greetingChannel) GreetRouter(natsClient diegonats.NATSClient) {
	greetingMsg, err := natsClient.Request(RouterGreetTopic, []byte{}, NatsGreetRequestTimeout)
	if err != nil {
		return
	}

	c.SendGreeting(greetingMsg)
}

func (c greetingChannel) SendGreeting(greetingMsg *nats.Msg) {
	greeting := &GreetingMessage{}

	err := json.Unmarshal(greetingMsg.Data, greeting)
	if err != nil {
		return
	}

	select {
	case c <- greeting:
	default:
	}
}
