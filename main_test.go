package main_test

import (
	"fmt"
	"net/http"

	. "github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/api"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"
	"github.com/tedsuo/rata"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/ginkgo/testrunner"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var receptorBinPath
var receptorURL string

var _ = SynchronizedBeforeSuite(
	func() []byte {
		receptorConfig, err := gexec.Build("github.com/cloudfoundry-incubator/receptor", "-race")
		Ω(err).ShouldNot(HaveOccurred())
		return []byte(receptorConfig)
	},
	func(receptorConfig []byte) {
		receptorBinPath = string(receptorConfig)
		receptorUrl = fmt.Sprintf("127.0.0.1:%d", 6700+GinkgoParallelNode)
	},
)

var _ = SynchronizedAfterSuite(func() {
}, func() {
	gexec.CleanupBuildArtifacts()
})

var _ = Describe("Receptor API", func() {
	var receptorRunner *ginkgomon.Runner
	var receptorProcess ifrit.Process
	var reqGen *rata.RequestGenerator
	var client *http.Client

	BeforeEach(func() {
		receptorRunner = testrunner.New(receptorBinPath, receptorURL)
		receptorProcess = ginkgomon.Invoke(receptorRunner)
		reqGen = rata.NewRequestGenerator(receptorURL, api.Routes)
		client = new(http.Client)
	})

	AfterEach(func() {
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
				Actions: []models.ExecutorAction{
					{Action: models.RunAction{Path: "/bin/bash", Args: []string{"echo", "hi"}}},
				},
			}
			createTaskReq, err := reqGen.CreateRequest(api.CreateTask, nil, taskToCreate.JSONReader())
			Ω(err).ShouldNot(HaveOccurred())
			createTaskRes, err = client.Do(createTaskReq)
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("responds with 201 CREATED", func() {})
		It("desires the task in the BBS", func() {})
	})
})
