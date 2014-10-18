package main_test

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/cmd/receptor/testrunner"
	"github.com/cloudfoundry-incubator/receptor/handlers"
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

const username = "username"
const password = "password"

var receptorBinPath string
var receptorAddress string
var etcdPort int

var _ = SynchronizedBeforeSuite(
	func() []byte {
		receptorConfig, err := gexec.Build("github.com/cloudfoundry-incubator/receptor/cmd/receptor", "-race")
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
	var etcdUrl string
	var etcdRunner *etcdstorerunner.ETCDClusterRunner
	var bbs *Bbs.BBS
	var receptorRunner *ginkgomon.Runner
	var receptorProcess ifrit.Process
	var reqGen *rata.RequestGenerator
	var client *http.Client

	BeforeEach(func() {
		etcdUrl = fmt.Sprintf("http://127.0.0.1:%d", etcdPort)
		etcdRunner = etcdstorerunner.NewETCDClusterRunner(etcdPort, 1)
		etcdRunner.Start()

		logger := lager.NewLogger("bbs")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))

		bbs = Bbs.NewBBS(etcdRunner.Adapter(), timeprovider.NewTimeProvider(), logger)

		reqGen = rata.NewRequestGenerator("http://"+receptorAddress, handlers.Routes)
		reqGen.Header.Set("Authorization", "Basic "+basicAuth(username, password))
		client = new(http.Client)

		receptorRunner = testrunner.New(receptorBinPath, receptorAddress, etcdUrl, username, password)
		receptorProcess = ginkgomon.Invoke(receptorRunner)
	})

	AfterEach(func() {
		defer etcdRunner.Stop()
		ginkgomon.Kill(receptorProcess)
	})

	Describe("Basic Auth", func() {
		var res *http.Response

		Context("when the username and password are blank", func() {
			BeforeEach(func() {
				var err error
				ginkgomon.Kill(receptorProcess)
				receptorRunner = testrunner.New(receptorBinPath, receptorAddress, etcdUrl, "", "")
				receptorProcess = ginkgomon.Invoke(receptorRunner)
				res, err = client.Get("http://" + receptorAddress)
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
				res, err = client.Get("http://" + receptorAddress)
				Ω(err).ShouldNot(HaveOccurred())
				res.Body.Close()
			})

			It("returns 401 for all requests", func() {
				Ω(res.StatusCode).Should(Equal(http.StatusUnauthorized))
			})
		})
	})

	Describe("POST /task", func() {
		var createTaskReq *http.Request
		var createTaskRes *http.Response
		var taskToCreate receptor.CreateTaskRequest

		BeforeEach(func() {
			taskToCreate = receptor.CreateTaskRequest{
				TaskGuid: "task-guid-1",
				Domain:   "test-domain",
				Stack:    "some-stack",
				Actions: []models.ExecutorAction{
					{Action: models.RunAction{Path: "/bin/bash", Args: []string{"echo", "hi"}}},
				},
			}
			var err error
			createTaskReq, err = reqGen.CreateRequest(handlers.CreateTask, nil, taskToCreate.JSONReader())
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
				createTaskReq, err = reqGen.CreateRequest(handlers.CreateTask, nil, taskToCreate.JSONReader())
				Ω(err).ShouldNot(HaveOccurred())

				createTaskRes, err = client.Do(createTaskReq)
				Ω(err).ShouldNot(HaveOccurred())
				body, err = ioutil.ReadAll(createTaskRes.Body)
				Ω(err).ShouldNot(HaveOccurred())
				createTaskRes.Body.Close()
			})

			It("returns an error indicating that the key already exists", func() {
				Ω(createTaskRes.StatusCode).Should(Equal(http.StatusInternalServerError))

				expectedError := receptor.ErrorResponse{
					Error: storeadapter.ErrorKeyExists.Error(),
				}

				Expect(body).To(BeEquivalentTo(expectedError.JSONReader().String()))
			})
		})
	})
})

// See 2 (end of page 4) http://www.ietf.org/rfc/rfc2617.txt
// "To receive authorization, the client sends the userid and password,
// separated by a single colon (":") character, within a base64
// encoded string in the credentials."
// It is not meant to be urlencoded.
func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
