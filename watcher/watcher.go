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

			w.handleDesiredCreate(logger, desiredCreate)

		case desiredUpdate, ok := <-desiredLRPUpdates:
			if !ok {
				logger.Info("desired-lrp-update-channel-closed")
				desiredLRPCreates = nil
				desiredLRPUpdates = nil
				desiredLRPDeletes = nil
				break
			}

			w.handleDesiredUpdate(logger, desiredUpdate)

		case desiredDelete, ok := <-desiredLRPDeletes:
			if !ok {
				logger.Info("desired-lrp-delete-channel-closed")
				desiredLRPCreates = nil
				desiredLRPUpdates = nil
				desiredLRPDeletes = nil
				break
			}

			w.handleDesiredDelete(logger, desiredDelete)

		case actualCreate, ok := <-actualLRPCreates:
			if !ok {
				logger.Info("actual-lrp-create-channel-closed")
				actualLRPCreates = nil
				actualLRPUpdates = nil
				actualLRPDeletes = nil
				break
			}

			w.handleActualCreate(logger, actualCreate)

		case actualUpdate, ok := <-actualLRPUpdates:
			if !ok {
				logger.Info("actual-lrp-update-channel-closed")
				actualLRPCreates = nil
				actualLRPUpdates = nil
				actualLRPDeletes = nil
				break
			}

			w.handleActualUpdate(logger, actualUpdate)

		case actualDelete, ok := <-actualLRPDeletes:
			if !ok {
				logger.Info("actual-lrp-delete-channel-closed")
				actualLRPCreates = nil
				actualLRPUpdates = nil
				actualLRPDeletes = nil
				break
			}

			w.handleActualDelete(logger, actualDelete)

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

func (w *watcher) handleDesiredCreate(logger lager.Logger, desiredLrp models.DesiredLRP) {
	logger.Info("handling-desired-create", lager.Data{"desired-lrp": desiredLrp})
	defer logger.Info("done-handling-desired-create")

	logger.Info("emitting-desired-created-event", lager.Data{"desired-lrp": desiredLrp})
	w.hub.Emit(receptor.NewDesiredLRPCreatedEvent(serialization.DesiredLRPToResponse(desiredLrp)))
}

func (w *watcher) handleDesiredUpdate(logger lager.Logger, desiredLRPChange models.DesiredLRPChange) {
	before, after := desiredLRPChange.Before, desiredLRPChange.After

	logger.Info("handling-desired-change", lager.Data{"before": before, "after": after})
	defer logger.Info("done-handling-desired-update")

	logger.Info("emitting-desired-updated-event", lager.Data{"before": before, "after": after})
	w.hub.Emit(receptor.NewDesiredLRPChangedEvent(
		serialization.DesiredLRPToResponse(before),
		serialization.DesiredLRPToResponse(after),
	))
}

func (w *watcher) handleDesiredDelete(logger lager.Logger, desiredLrp models.DesiredLRP) {
	logger.Info("handling-desired-delete", lager.Data{"desired-lrp": desiredLrp})
	defer logger.Info("done-handling-desired-delete")

	logger.Info("emitting-desired-removed-event", lager.Data{"desired-lrp": desiredLrp})
	w.hub.Emit(receptor.NewDesiredLRPRemovedEvent(serialization.DesiredLRPToResponse(desiredLrp)))
}

func (w *watcher) handleActualCreate(logger lager.Logger, actualLrp models.ActualLRP) {
	logger.Info("handling-actual-create", lager.Data{"actual-lrp": actualLrp})
	defer logger.Info("done-handling-actual-create")

	logger.Info("emitting-actual-created-event", lager.Data{"actual-lrp": actualLrp})
	w.hub.Emit(receptor.NewActualLRPCreatedEvent(serialization.ActualLRPToResponse(actualLrp)))
}

func (w *watcher) handleActualUpdate(logger lager.Logger, actualLRPChange models.ActualLRPChange) {
	before, after := actualLRPChange.Before, actualLRPChange.After

	logger.Info("handling-actual-change", lager.Data{"before": before, "after": after})
	defer logger.Info("done-handling-actual-update")

	logger.Info("emitting-actual-updated-event", lager.Data{"before": before, "after": after})
	w.hub.Emit(receptor.NewActualLRPChangedEvent(
		serialization.ActualLRPToResponse(before),
		serialization.ActualLRPToResponse(after),
	))
}

func (w *watcher) handleActualDelete(logger lager.Logger, actualLrp models.ActualLRP) {
	logger.Info("handling-actual-delete", lager.Data{"actual-lrp": actualLrp})
	defer logger.Info("done-handling-actual-delete")

	logger.Info("emitting-actual-removed-event", lager.Data{"actual-lrp": actualLrp})
	w.hub.Emit(receptor.NewActualLRPRemovedEvent(serialization.ActualLRPToResponse(actualLrp)))
}
