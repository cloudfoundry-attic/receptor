package watcher

import (
	"os"
	"time"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/event"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	"github.com/cloudfoundry-incubator/runtime-schema/bbs"
	"github.com/cloudfoundry/gunk/timeprovider"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"
)

type Watcher ifrit.Runner

type watcher struct {
	bbs                     bbs.ReceptorBBS
	hub                     event.Hub
	timeProvider            timeprovider.TimeProvider
	retryWaitDuration       time.Duration
	subscriberCheckDuration time.Duration
	logger                  lager.Logger
}

func NewWatcher(
	bbs bbs.ReceptorBBS,
	hub event.Hub,
	timeProvider timeprovider.TimeProvider,
	retryWaitDuration time.Duration,
	subscriberCheckDuration time.Duration,
	logger lager.Logger,
) Watcher {
	return &watcher{
		bbs:                     bbs,
		hub:                     hub,
		timeProvider:            timeProvider,
		retryWaitDuration:       retryWaitDuration,
		subscriberCheckDuration: subscriberCheckDuration,
		logger:                  logger,
	}
}

func (w *watcher) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	logger := w.logger.Session("running-watcher")
	logger.Info("starting")

	desiredLRPCreates, desiredLRPUpdates, desiredLRPDeletes, desiredErrors := w.bbs.WatchForDesiredLRPChanges(logger)
	actualLRPCreates, actualLRPUpdates, actualLRPDeletes, actualErrors := w.bbs.WatchForActualLRPChanges(logger)

	close(ready)
	logger.Info("started")
	defer logger.Info("finished")

	var reWatchActual <-chan time.Time
	var reWatchDesired <-chan time.Time
	subscriberCheckTicker := w.timeProvider.NewTicker(w.subscriberCheckDuration)

	for {
		select {
		case desiredCreate, ok := <-desiredLRPCreates:
			if !ok {
				logger.Info("desired-lrp-create-channel-closed")
				desiredLRPCreates = nil
				desiredLRPUpdates = nil
				desiredLRPDeletes = nil
				break
			}

			logger.Debug("received-desired-lrp-create-event")
			w.hub.Emit(receptor.NewDesiredLRPCreatedEvent(serialization.DesiredLRPToResponse(desiredCreate)))

		case desiredUpdate, ok := <-desiredLRPUpdates:
			if !ok {
				logger.Info("desired-lrp-update-channel-closed")
				desiredLRPCreates = nil
				desiredLRPUpdates = nil
				desiredLRPDeletes = nil
				break
			}

			logger.Debug("received-desired-lrp-update-event")

			before, after := desiredUpdate.Before, desiredUpdate.After
			w.hub.Emit(receptor.NewDesiredLRPChangedEvent(
				serialization.DesiredLRPToResponse(before),
				serialization.DesiredLRPToResponse(after),
			))

		case desiredDelete, ok := <-desiredLRPDeletes:
			if !ok {
				logger.Info("desired-lrp-delete-channel-closed")
				desiredLRPCreates = nil
				desiredLRPUpdates = nil
				desiredLRPDeletes = nil
				break
			}

			logger.Debug("received-desired-lrp-delete-event")
			w.hub.Emit(receptor.NewDesiredLRPRemovedEvent(serialization.DesiredLRPToResponse(desiredDelete)))

		case actualCreate, ok := <-actualLRPCreates:
			if !ok {
				logger.Info("actual-lrp-create-channel-closed")
				actualLRPCreates = nil
				actualLRPUpdates = nil
				actualLRPDeletes = nil
				break
			}

			logger.Debug("received-actual-lrp-create-event")
			w.hub.Emit(receptor.NewActualLRPCreatedEvent(serialization.ActualLRPToResponse(actualCreate)))

		case actualUpdate, ok := <-actualLRPUpdates:
			if !ok {
				logger.Info("actual-lrp-update-channel-closed")
				actualLRPCreates = nil
				actualLRPUpdates = nil
				actualLRPDeletes = nil
				break
			}

			logger.Debug("received-actual-lrp-update-event")

			before, after := actualUpdate.Before, actualUpdate.After
			w.hub.Emit(receptor.NewActualLRPChangedEvent(
				serialization.ActualLRPToResponse(before),
				serialization.ActualLRPToResponse(after),
			))

		case actualDelete, ok := <-actualLRPDeletes:
			if !ok {
				logger.Info("actual-lrp-delete-channel-closed")
				actualLRPCreates = nil
				actualLRPUpdates = nil
				actualLRPDeletes = nil
				break
			}

			logger.Debug("received-actual-lrp-delete-event")
			w.hub.Emit(receptor.NewActualLRPRemovedEvent(serialization.ActualLRPToResponse(actualDelete)))

		case err := <-desiredErrors:
			logger.Error("desired-watch-failed", err)

			reWatchDesired = w.timeProvider.NewTimer(w.retryWaitDuration).C()
			desiredLRPCreates = nil
			desiredLRPUpdates = nil
			desiredLRPDeletes = nil
			desiredErrors = nil

		case err := <-actualErrors:
			logger.Error("actual-watch-failed", err)

			reWatchActual = w.timeProvider.NewTimer(w.retryWaitDuration).C()
			actualLRPCreates = nil
			actualLRPUpdates = nil
			actualLRPDeletes = nil
			actualErrors = nil

		case <-reWatchDesired:
			desiredLRPCreates, desiredLRPUpdates, desiredLRPDeletes, desiredErrors = w.bbs.WatchForDesiredLRPChanges(logger)
			reWatchDesired = nil

		case <-reWatchActual:
			actualLRPCreates, actualLRPUpdates, actualLRPDeletes, actualErrors = w.bbs.WatchForActualLRPChanges(logger)
			reWatchActual = nil

		case <-subscriberCheckTicker.C():
			if w.hub.HasSubscribers() {
				desiredLRPCreates, desiredLRPUpdates, desiredLRPDeletes, desiredErrors = w.bbs.WatchForDesiredLRPChanges(logger)
				actualLRPCreates, actualLRPUpdates, actualLRPDeletes, actualErrors = w.bbs.WatchForActualLRPChanges(logger)
			} else {
				desiredLRPCreates = nil
				desiredLRPUpdates = nil
				desiredLRPDeletes = nil
				desiredErrors = nil
				actualLRPCreates = nil
				actualLRPUpdates = nil
				actualLRPDeletes = nil
				actualErrors = nil
			}

		case <-signals:
			logger.Info("stopping")
			subscriberCheckTicker.Stop()
			return nil
		}
	}

	return nil
}
