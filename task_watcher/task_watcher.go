package task_watcher

import (
	"errors"
	"os"

	Bbs "github.com/cloudfoundry-incubator/runtime-schema/bbs"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"
)

const MAX_RETRIES = 3
const POOL_SIZE = 20

var ErrWatchFailed = errors.New("watching for completed tasks failed")

func New(bbs Bbs.ReceptorBBS, logger lager.Logger) *TaskWatcher {
	return &TaskWatcher{
		bbs:    bbs,
		logger: logger.Session("task-watcher"),
	}
}

type TaskWatcher struct {
	bbs    Bbs.ReceptorBBS
	logger lager.Logger
}

func (t *TaskWatcher) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	t.logger.Info("starting")

	workQueue := make(chan models.Task, POOL_SIZE)
	workers := ifrit.Invoke(newTaskWorkerPool(POOL_SIZE, workQueue, t.bbs, t.logger))
	workersDone := workers.Wait()

	close(ready)

	taskChan, stopChan, errChan := t.bbs.WatchForCompletedTask()
	t.logger.Info("watching")

	for {
		select {
		case task, ok := <-taskChan:
			if !ok {
				taskChan = nil
				break
			}
			select {
			case workQueue <- task:
			default:
				t.logger.Info("queue-full-ignoring-task", lager.Data{"task-guid": task.TaskGuid})
			}

		case sig := <-signals:
			t.logger.Info("stopping")
			workers.Signal(sig)
			if sig == os.Kill {
				return nil
			}
			stopChan <- true

		case <-workersDone:
			t.logger.Info("exited")
			return nil

		case err, ok := <-errChan:
			if !ok {
				errChan = nil
				break
			}
			t.logger.Error("watch-failed", err)
			taskChan, stopChan, errChan = t.bbs.WatchForCompletedTask()
			t.logger.Info("watching-again")
		}
	}
}
