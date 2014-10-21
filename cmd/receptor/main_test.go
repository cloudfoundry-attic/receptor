package main_test

import (
	"fmt"
	"net/http"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/cmd/receptor/testrunner"
	Bbs "github.com/cloudfoundry-incubator/runtime-schema/bbs"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/cloudfoundry/gunk/timeprovider"
	"github.com/cloudfoundry/storeadapter/storerunner/etcdstorerunner"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Receptor API", func() {
	var etcdUrl string
	var etcdRunner *etcdstorerunner.ETCDClusterRunner
	var bbs *Bbs.BBS
	var receptorRunner *ginkgomon.Runner
	var receptorProcess ifrit.Process
	var client receptor.Client

	BeforeEach(func() {
		etcdUrl = fmt.Sprintf("http://127.0.0.1:%d", etcdPort)
		etcdRunner = etcdstorerunner.NewETCDClusterRunner(etcdPort, 1)
		etcdRunner.Start()

		logger := lager.NewLogger("bbs")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))

		bbs = Bbs.NewBBS(etcdRunner.Adapter(), timeprovider.NewTimeProvider(), logger)

		client = receptor.NewClient(receptorAddress, username, password)

		receptorRunner = testrunner.New(receptorBinPath, receptorAddress, etcdUrl, username, password)
		receptorProcess = ginkgomon.Invoke(receptorRunner)
	})

	AfterEach(func() {
		defer etcdRunner.Stop()
		ginkgomon.Kill(receptorProcess)
	})

	Describe("Basic Auth", func() {
		var res *http.Response
		var httpClient *http.Client

		BeforeEach(func() {
			httpClient = new(http.Client)
		})

		Context("when the username and password are blank", func() {
			BeforeEach(func() {
				var err error
				ginkgomon.Kill(receptorProcess)
				receptorRunner = testrunner.New(receptorBinPath, receptorAddress, etcdUrl, "", "")
				receptorProcess = ginkgomon.Invoke(receptorRunner)

				res, err = httpClient.Get("http://" + receptorAddress)
				Ω(err).ShouldNot(HaveOccurred())
				res.Body.Close()
			})

			It("does not return 401", func() {
				Ω(res.StatusCode).Should(Equal(http.StatusNotFound))
			})
		})

		Context("when the username and password are required but not sent", func() {
			BeforeEach(func() {
				var err error
				res, err = httpClient.Get("http://" + receptorAddress)
				Ω(err).ShouldNot(HaveOccurred())
				res.Body.Close()
			})

			It("returns 401 for all requests", func() {
				Ω(res.StatusCode).Should(Equal(http.StatusUnauthorized))
			})
		})
	})

	Describe("Headers", func() {
		It("includes the Content-Length and Content-Type headers", func() {
			httpClient := new(http.Client)
			res, err := httpClient.Get("http://" + receptorAddress + "/tasks")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(res.Header.Get("Content-Length")).Should(MatchRegexp(`\d+`))
			Ω(res.Header.Get("Content-Type")).Should(Equal("application/json"))
		})
	})

	Describe("POST /tasks", func() {
		var taskToCreate receptor.CreateTaskRequest
		var err error

		BeforeEach(func() {
			taskToCreate = receptor.CreateTaskRequest{
				TaskGuid: "task-guid-1",
				Domain:   "test-domain",
				Stack:    "some-stack",
				Actions: []models.ExecutorAction{
					{Action: models.RunAction{Path: "/bin/bash", Args: []string{"echo", "hi"}}},
				},
			}

			err = client.CreateTask(taskToCreate)
		})

		It("responds without an error", func() {
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("desires the task in the BBS", func() {
			Eventually(bbs.GetAllPendingTasks).Should(HaveLen(1))
		})

		Context("when trying to create a task with a GUID that already exists", func() {
			BeforeEach(func() {
				err = client.CreateTask(taskToCreate)
			})

			It("returns an error indicating that the key already exists", func() {
				Ω(err.(receptor.Error).Type).Should(Equal(receptor.TaskGuidAlreadyExists))
			})
		})
	})

	Describe("GET /tasks", func() {
		Context("when there are no tasks", func() {
			It("returns an empty array", func() {
				tasks, err := client.GetAllTasks()
				Ω(err).ShouldNot(HaveOccurred())
				Ω(tasks).Should(BeEmpty())
			})
		})

		Context("when there are tasks", func() {
			BeforeEach(func() {
				err := bbs.DesireTask(models.Task{
					TaskGuid: "task-guid-1",
					Domain:   "test-domain",
					Stack:    "some-stack",
					Actions: []models.ExecutorAction{
						{Action: models.RunAction{Path: "/bin/true"}},
					},
				})
				Ω(err).ShouldNot(HaveOccurred())

				err = bbs.DesireTask(models.Task{
					TaskGuid: "task-guid-2",
					Domain:   "test-domain",
					Stack:    "some-stack",
					Actions: []models.ExecutorAction{
						{Action: models.RunAction{Path: "/bin/true"}},
					},
				})
				Ω(err).ShouldNot(HaveOccurred())
			})

			It("returns an array of all the tasks", func() {
				tasks, err := client.GetAllTasks()
				Ω(err).ShouldNot(HaveOccurred())

				taskGuids := []string{}
				for _, task := range tasks {
					taskGuids = append(taskGuids, task.TaskGuid)
				}
				Ω(taskGuids).Should(ConsistOf([]string{"task-guid-1", "task-guid-2"}))
			})
		})
	})

	Describe("GET /domains/:domain/tasks", func() {
		BeforeEach(func() {
			err := bbs.DesireTask(models.Task{
				TaskGuid: "task-guid-1",
				Domain:   "test-domain",
				Stack:    "stack-1",
				Actions: []models.ExecutorAction{
					{Action: models.RunAction{Path: "/bin/true"}},
				},
			})
			Ω(err).ShouldNot(HaveOccurred())

			err = bbs.DesireTask(models.Task{
				TaskGuid: "task-guid-2",
				Domain:   "other-domain",
				Stack:    "stack-2",
				Actions: []models.ExecutorAction{
					{Action: models.RunAction{Path: "/bin/true"}},
				},
			})
			Ω(err).ShouldNot(HaveOccurred())

			err = bbs.DesireTask(models.Task{
				TaskGuid: "task-guid-3",
				Domain:   "test-domain",
				Stack:    "stack-3",
				Actions: []models.ExecutorAction{
					{Action: models.RunAction{Path: "/bin/true"}},
				},
			})
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("returns an array of all the tasks for the domain", func() {
			tasks, err := client.GetAllTasksByDomain("test-domain")
			Ω(err).ShouldNot(HaveOccurred())

			taskGuids := []string{}
			for _, task := range tasks {
				taskGuids = append(taskGuids, task.TaskGuid)
			}
			Ω(taskGuids).Should(ConsistOf([]string{"task-guid-1", "task-guid-3"}))
		})
	})
})
