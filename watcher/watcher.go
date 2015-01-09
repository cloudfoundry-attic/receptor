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
	desiredLRPChanges, _, desiredErrors := w.bbs.WatchForDesiredLRPChanges()
	actualLRPChanges, _, actualErrors := w.bbs.WatchForActualLRPChanges()

	close(ready)
	logger.Info("started")
	defer logger.Info("finished")

	var reWatchActual <-chan time.Time
	var reWatchDesired <-chan time.Time

	for {
		select {
		case desiredChange, ok := <-desiredLRPChanges:
			if !ok {
				logger.Info("desired-lrp-channel-closed")
				desiredLRPChanges = nil
				break
			}

			w.handleDesiredChange(logger, desiredChange)

		case actualChange, ok := <-actualLRPChanges:
			if !ok {
				logger.Info("actual-lrp-channel-closed")
				actualLRPChanges = nil
				break
			}

			w.handleActualChange(logger, actualChange)

		case err := <-desiredErrors:
			logger.Error("desired-watch-failed", err)

			reWatchDesired = w.timeProvider.NewTimer(w.retryWaitDuration).C()
			desiredLRPChanges = nil
			desiredErrors = nil

		case err := <-actualErrors:
			logger.Error("actual-watch-failed", err)

			reWatchActual = w.timeProvider.NewTimer(w.retryWaitDuration).C()
			actualLRPChanges = nil
			actualErrors = nil

		case <-reWatchActual:
			actualLRPChanges, _, actualErrors = w.bbs.WatchForActualLRPChanges()
			reWatchActual = nil

		case <-reWatchDesired:
			desiredLRPChanges, _, desiredErrors = w.bbs.WatchForDesiredLRPChanges()
			reWatchDesired = nil

		case <-signals:
			logger.Info("stopping")
			return nil
		}
	}

	return nil
}

func (w *watcher) handleDesiredChange(logger lager.Logger, change models.DesiredLRPChange) {
	logger.Info("handling-desired-change", lager.Data{"change": change})
	defer logger.Info("done-handling-desired-change")
	before := change.Before
	after := change.After

	switch {
	case after == nil && before == nil:
		logger.Debug("received-invalid-desiredLRP-change")
	case after == nil && before != nil:
		logger.Info("emitting-desired-remove-event", lager.Data{"process-guid": before.ProcessGuid})
		w.hub.Emit(receptor.NewDesiredLRPRemovedEvent(before.ProcessGuid))
	default:
		logger.Info("emitting-desired-change-event", lager.Data{"response": *after})
		w.hub.Emit(receptor.NewDesiredLRPChangedEvent(serialization.DesiredLRPToResponse(*after)))
	}
}

func (w *watcher) handleActualChange(logger lager.Logger, change models.ActualLRPChange) {
	logger.Info("handling-actual-change", lager.Data{"change": change})
	defer logger.Info("done-handling-actual-change")
	before := change.Before
	after := change.After

	switch {
	case after == nil && before == nil:
		logger.Debug("received-invalid-actualLRP-change")
	case after == nil && before != nil:
		logger.Info("emitting-actual-remove-event", lager.Data{"process-guid": before.ProcessGuid, "index": before.Index})
		w.hub.Emit(receptor.NewActualLRPRemovedEvent(before.ProcessGuid, before.Index))
	default:
		logger.Info("emitting-actual-change-event", lager.Data{"response": *after})
		w.hub.Emit(receptor.NewActualLRPChangedEvent(serialization.ActualLRPToResponse(*after)))
	}
}
