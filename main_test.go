package main_test

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/cloudfoundry-incubator/receptor/api"
	"github.com/cloudfoundry-incubator/receptor/testrunner"
	Bbs "github.com/cloudfoundry-incubator/runtime-schema/bbs"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/cloudfoundry/gunk/timeprovider"
	"github.com/cloudfoundry/storeadapter"
	"github.com/cloudfoundry/storeadapter/storerunner/etcdstorerunner"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"
	"github.com/tedsuo/rata"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var receptorBinPath string
var receptorAddress string
var etcdPort int

var _ = SynchronizedBeforeSuite(
	func() []byte {
		receptorConfig, err := gexec.Build("github.com/cloudfoundry-incubator/receptor", "-race")
		Ω(err).ShouldNot(HaveOccurred())
		return []byte(receptorConfig)
	},
	func(receptorConfig []byte) {
		receptorBinPath = string(receptorConfig)
		receptorAddress = fmt.Sprintf("127.0.0.1:%d", 6700+GinkgoParallelNode())
		etcdPort = 4001 + GinkgoParallelNode()
	},
)

var _ = SynchronizedAfterSuite(func() {
}, func() {
	gexec.CleanupBuildArtifacts()
})

var _ = Describe("Receptor API", func() {
	var etcdRunner *etcdstorerunner.ETCDClusterRunner
	var bbs *Bbs.BBS
	var receptorRunner *ginkgomon.Runner
	var receptorProcess ifrit.Process
	var reqGen *rata.RequestGenerator
	var client *http.Client

	BeforeEach(func() {
		etcdRunner = etcdstorerunner.NewETCDClusterRunner(etcdPort, 1)
		etcdRunner.Start()
		logger := lager.NewLogger("bbs")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
		bbs = Bbs.NewBBS(etcdRunner.Adapter(), timeprovider.NewTimeProvider(), logger)

		etcdUrl := fmt.Sprintf("http://127.0.0.1:%d", etcdPort)
		receptorRunner = testrunner.New(receptorBinPath, receptorAddress, etcdUrl)
		receptorProcess = ginkgomon.Invoke(receptorRunner)
		reqGen = rata.NewRequestGenerator("http://"+receptorAddress, api.Routes)
		client = new(http.Client)
	})

	AfterEach(func() {
		defer etcdRunner.Stop()
		ginkgomon.Kill(receptorProcess)
	})

	Describe("POST /task", func() {
		var createTaskReq *http.Request
		var createTaskRes *http.Response
		var taskToCreate api.CreateTaskRequest

		BeforeEach(func() {
			taskToCreate = api.CreateTaskRequest{
				TaskGuid: "task-guid-1",
				Domain:   "test-domain",
				Stack:    "some-stack",
				Actions: []models.ExecutorAction{
					{Action: models.RunAction{Path: "/bin/bash", Args: []string{"echo", "hi"}}},
				},
			}
			var err error
			createTaskReq, err = reqGen.CreateRequest(api.CreateTask, nil, taskToCreate.JSONReader())

			Ω(err).ShouldNot(HaveOccurred())
			createTaskRes, err = client.Do(createTaskReq)
			Ω(err).ShouldNot(HaveOccurred())
			createTaskRes.Body.Close()
		})

		It("responds with 201 CREATED", func() {
			Ω(createTaskRes.StatusCode).Should(Equal(http.StatusCreated))
		})

		It("desires the task in the BBS", func() {
			Eventually(bbs.GetAllPendingTasks).Should(HaveLen(1))
		})

		Context("when trying to create a task with a GUID that already exists", func() {
			var body []byte

			BeforeEach(func() {
				var err error
				createTaskReq, err = reqGen.CreateRequest(api.CreateTask, nil, taskToCreate.JSONReader())
				Ω(err).ShouldNot(HaveOccurred())

				createTaskRes, err = client.Do(createTaskReq)
				Ω(err).ShouldNot(HaveOccurred())
				body, err = ioutil.ReadAll(createTaskRes.Body)
				Ω(err).ShouldNot(HaveOccurred())
				createTaskRes.Body.Close()
			})

			It("returns an error indicating that the key already exists", func() {
				Ω(createTaskRes.StatusCode).Should(Equal(http.StatusInternalServerError))

				expectedError := api.ErrorResponse{
					Error: storeadapter.ErrorKeyExists.Error(),
				}

				Expect(body).To(BeEquivalentTo(expectedError.JSONReader().String()))
			})
		})
	})
})
