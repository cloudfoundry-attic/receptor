package heartbeat

import (
	"os"
	"time"

	"github.com/cloudfoundry/gunk/diegonats"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/restart"
)

func NewBackgroundHeartbeat(natsAddress, natsUsername, natsPassword string, logger lager.Logger, registration RegistryMessage) ifrit.RunFunc {
	return func(signals <-chan os.Signal, ready chan<- struct{}) error {
		restarter := restart.Restarter{
			Runner: newBackgroundGroup(natsAddress, natsUsername, natsPassword, logger, registration),
			Load: func(runner ifrit.Runner, err error) ifrit.Runner {
				return newBackgroundGroup(natsAddress, natsUsername, natsPassword, logger, registration)
			},
		}
		// don't wait, start this thing in the background
		close(ready)
		return restarter.Run(signals, make(chan struct{}))
	}
}

func newBackgroundGroup(natsAddress, natsUsername, natsPassword string, logger lager.Logger, registration RegistryMessage) grouper.StaticGroup {
	client := diegonats.NewClient()
	return grouper.NewOrdered(os.Interrupt, grouper.Members{
		{"nats_connection", diegonats.NewClientRunner(natsAddress, natsUsername, natsPassword, logger, client)},
		{"router_heartbeat", New(client, registration, 50*time.Millisecond, logger)},
	})
}
