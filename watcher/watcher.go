package watcher

import (
	"os"
	"time"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/event"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	"github.com/cloudfoundry-incubator/runtime-schema/bbs"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/cloudfoundry/gunk/timeprovider"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"
)

type Watcher ifrit.Runner

type watcher struct {
	bbs               bbs.ReceptorBBS
	hub               event.Hub
	timeProvider      timeprovider.TimeProvider
	retryWaitDuration time.Duration
	logger            lager.Logger
}

func NewWatcher(
	bbs bbs.ReceptorBBS,
	hub event.Hub,
	timeProvider timeprovider.TimeProvider,
	retryWaitDuration time.Duration,
	logger lager.Logger,
) Watcher {
	return &watcher{
		bbs:               bbs,
		hub:               hub,
		timeProvider:      timeProvider,
		retryWaitDuration: retryWaitDuration,
		logger:            logger,
	}
}

func (w *watcher) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	logger := w.logger.Session("running-watcher")
	logger.Info("starting")
	desiredLRPCreateOrUpdates, desiredLRPDeletes, desiredErrors := w.bbs.WatchForDesiredLRPChanges(logger)
	actualLRPCreateOrUpdates, actualLRPDeletes, actualErrors := w.bbs.WatchForActualLRPChanges(logger)

	close(ready)
	logger.Info("started")
	defer logger.Info("finished")

	var reWatchActual <-chan time.Time
	var reWatchDesired <-chan time.Time

	for {
		select {
		case desiredCreateOrUpdate, ok := <-desiredLRPCreateOrUpdates:
			if !ok {
				logger.Info("desired-lrp-create-or-update-channel-closed")
				desiredLRPCreateOrUpdates = nil
				desiredLRPDeletes = nil
				break
			}

			w.handleDesiredCreateOrUpdate(logger, desiredCreateOrUpdate)

		case desiredDelete, ok := <-desiredLRPDeletes:
			if !ok {
				logger.Info("desired-lrp-delete-channel-closed")
				desiredLRPCreateOrUpdates = nil
				desiredLRPDeletes = nil
				break
			}

			w.handleDesiredDelete(logger, desiredDelete)

		case actualCreateOrUpdate, ok := <-actualLRPCreateOrUpdates:
			if !ok {
				logger.Info("actual-lrp-create-or-update-channel-closed")
				actualLRPCreateOrUpdates = nil
				actualLRPDeletes = nil
				break
			}

			w.handleActualCreateOrUpdate(logger, actualCreateOrUpdate)

		case actualDelete, ok := <-actualLRPDeletes:
			if !ok {
				logger.Info("actual-lrp-delete-channel-closed")
				actualLRPCreateOrUpdates = nil
				actualLRPDeletes = nil
				break
			}

			w.handleActualDelete(logger, actualDelete)

		case err := <-desiredErrors:
			logger.Error("desired-watch-failed", err)

			reWatchDesired = w.timeProvider.NewTimer(w.retryWaitDuration).C()
			desiredLRPCreateOrUpdates = nil
			desiredLRPDeletes = nil
			desiredErrors = nil

		case err := <-actualErrors:
			logger.Error("actual-watch-failed", err)

			reWatchActual = w.timeProvider.NewTimer(w.retryWaitDuration).C()
			actualLRPCreateOrUpdates = nil
			actualLRPDeletes = nil
			actualErrors = nil

		case <-reWatchDesired:
			desiredLRPCreateOrUpdates, desiredLRPDeletes, desiredErrors = w.bbs.WatchForDesiredLRPChanges(logger)
			reWatchDesired = nil

		case <-reWatchActual:
			actualLRPCreateOrUpdates, actualLRPDeletes, actualErrors = w.bbs.WatchForActualLRPChanges(logger)
			reWatchActual = nil

		case <-signals:
			logger.Info("stopping")
			return nil
		}
	}

	return nil
}

func (w *watcher) handleDesiredCreateOrUpdate(logger lager.Logger, desiredLrp models.DesiredLRP) {
	logger.Info("handling-desired-create-or-update", lager.Data{"desired-lrp": desiredLrp})
	defer logger.Info("done-handling-desired-create-or-update")

	logger.Info("emitting-desired-changed-event", lager.Data{"desired-lrp": desiredLrp})
	w.hub.Emit(receptor.NewDesiredLRPChangedEvent(serialization.DesiredLRPToResponse(desiredLrp)))
}

func (w *watcher) handleDesiredDelete(logger lager.Logger, desiredLrp models.DesiredLRP) {
	logger.Info("handling-desired-delete", lager.Data{"desired-lrp": desiredLrp})
	defer logger.Info("done-handling-desired-delete")

	logger.Info("emitting-desired-removed-event", lager.Data{"desired-lrp": desiredLrp})
	w.hub.Emit(receptor.NewDesiredLRPRemovedEvent(serialization.DesiredLRPToResponse(desiredLrp)))
}

func (w *watcher) handleActualCreateOrUpdate(logger lager.Logger, actualLrp models.ActualLRP) {
	logger.Info("handling-actual-create-or-update", lager.Data{"actual-lrp": actualLrp})
	defer logger.Info("done-handling-actual-create-or-update")

	logger.Info("emitting-actual-changed-event", lager.Data{"actual-lrp": actualLrp})
	w.hub.Emit(receptor.NewActualLRPChangedEvent(serialization.ActualLRPToResponse(actualLrp)))
}

func (w *watcher) handleActualDelete(logger lager.Logger, actualLrp models.ActualLRP) {
	logger.Info("handling-actual-delete", lager.Data{"actual-lrp": actualLrp})
	defer logger.Info("done-handling-actual-delete")

	logger.Info("emitting-actual-removed-event", lager.Data{"actual-lrp": actualLrp})
	w.hub.Emit(receptor.NewActualLRPRemovedEvent(serialization.ActualLRPToResponse(actualLrp)))
}
