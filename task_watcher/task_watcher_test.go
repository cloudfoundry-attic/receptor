package task_watcher_test

import (
	"errors"
	"net/http"
	"net/url"
	"os"

	"github.com/cloudfoundry-incubator/receptor/task_watcher"
	"github.com/cloudfoundry-incubator/runtime-schema/bbs/fake_bbs"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("TaskWatcher", func() {
	var (
		fakeBBS            *fake_bbs.FakeReceptorBBS
		taskWatcherProcess ifrit.Process
		taskChan           chan models.Task
		stopChan           chan bool
		errorChan          chan error
		err                error
		fakeServer         *ghttp.Server
	)

	BeforeEach(func() {
		fakeServer = ghttp.NewServer()

		taskChan = make(chan models.Task, 1)
		stopChan = make(chan bool, 1)
		errorChan = make(chan error, 1)

		logger := lager.NewLogger("task-watcher-test")

		fakeBBS = new(fake_bbs.FakeReceptorBBS)
		fakeBBS.WatchForCompletedTaskReturns(taskChan, stopChan, errorChan)

		taskWatcher := task_watcher.New(fakeBBS, logger)
		taskWatcherProcess = ifrit.Invoke(taskWatcher)
	})

	AfterEach(func() {
		taskWatcherProcess.Signal(os.Kill)
		Eventually(taskWatcherProcess.Wait()).Should(Receive())
	})

	Describe("shutting down", func() {
		It("stops watching for completed tasks", func() {
			taskWatcherProcess.Signal(os.Kill)
			Eventually(stopChan).Should(Receive())
		})
	})

	Describe("when a task is completed", func() {
		var (
			callbackURL *url.URL
			statusCodes chan int
		)

		BeforeEach(func() {
			statusCodes = make(chan int)

			fakeServer.RouteToHandler("POST", "/the-callback/url", func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(<-statusCodes)
			})

			callbackURL, err = url.Parse(fakeServer.URL() + "/the-callback/url")
			Ω(err).ShouldNot(HaveOccurred())
		})

		simulateTaskCompleting := func() {
			taskChan <- models.Task{
				TaskGuid:              "the-task-guid",
				CompletionCallbackURL: callbackURL,
			}
		}

		It("marks the task as resolving", func() {
			Ω(fakeBBS.ResolvingTaskCallCount()).Should(Equal(0))

			simulateTaskCompleting()
			statusCodes <- 200

			Eventually(fakeBBS.ResolvingTaskCallCount).Should(Equal(1))
			Ω(fakeBBS.ResolvingTaskArgsForCall(0)).Should(Equal("the-task-guid"))
		})

		Context("when marking the task as resolving fails", func() {
			BeforeEach(func() {
				fakeBBS.ResolvingTaskReturns(errors.New("failed to resolve task"))
			})

			It("does not make a request to the task's callback URL", func() {
				simulateTaskCompleting()

				Consistently(fakeServer.ReceivedRequests, 0.25).Should(BeEmpty())
			})
		})

		Context("when marking the task as resolving succeeds", func() {
			Context("when the task has a completion callback URL", func() {
				It("POSTs to the task's callback URL", func() {
					simulateTaskCompleting()

					statusCodes <- 200

					Eventually(fakeServer.ReceivedRequests).Should(HaveLen(1))
				})

				Context("when the request succeeds", func() {
					It("resolves the task", func() {
						simulateTaskCompleting()

						statusCodes <- 200

						Eventually(fakeBBS.ResolveTaskCallCount).Should(Equal(1))
						Ω(fakeBBS.ResolveTaskArgsForCall(0)).Should(Equal("the-task-guid"))
					})
				})

				Context("when the request fails with a 4xx response code", func() {
					It("resolves the task", func() {
						simulateTaskCompleting()

						statusCodes <- 403

						Eventually(fakeBBS.ResolveTaskCallCount).Should(Equal(1))
						Ω(fakeBBS.ResolveTaskArgsForCall(0)).Should(Equal("the-task-guid"))
					})
				})

				Context("when the request fails with a 5xx response code", func() {
					It("retries the request 2 more times", func() {
						simulateTaskCompleting()
						Eventually(fakeServer.ReceivedRequests).Should(HaveLen(1))

						statusCodes <- 502

						Consistently(fakeBBS.ResolveTaskCallCount, 0.25).Should(Equal(0))
						Eventually(fakeServer.ReceivedRequests).Should(HaveLen(2))

						statusCodes <- 502

						Consistently(fakeBBS.ResolveTaskCallCount, 0.25).Should(Equal(0))
						Eventually(fakeServer.ReceivedRequests).Should(HaveLen(3))

						statusCodes <- 200

						Eventually(fakeBBS.ResolveTaskCallCount, 0.25).Should(Equal(1))
						Ω(fakeBBS.ResolveTaskArgsForCall(0)).Should(Equal("the-task-guid"))
					})

					Context("when the request fails every time", func() {
						It("does not resolve the task", func() {
							simulateTaskCompleting()
							Eventually(fakeServer.ReceivedRequests).Should(HaveLen(1))

							statusCodes <- 502

							Consistently(fakeBBS.ResolveTaskCallCount, 0.25).Should(Equal(0))
							Eventually(fakeServer.ReceivedRequests).Should(HaveLen(2))

							statusCodes <- 502

							Consistently(fakeBBS.ResolveTaskCallCount, 0.25).Should(Equal(0))
							Eventually(fakeServer.ReceivedRequests).Should(HaveLen(3))

							statusCodes <- 502

							Consistently(fakeBBS.ResolveTaskCallCount, 0.25).Should(Equal(0))
							Consistently(fakeServer.ReceivedRequests, 0.25).Should(HaveLen(3))
						})
					})
				})
			})

			Context("when the task doesn't have a completion callback URL", func() {
				It("doesn't blow up", func() {
					callbackURL = nil
					simulateTaskCompleting()
				})
			})
		})
	})
})
