package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/cloudfoundry-incubator/receptor"
	. "github.com/cloudfoundry-incubator/receptor/handlers"
	"github.com/cloudfoundry-incubator/runtime-schema/bbs/fake_bbs"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/pivotal-golang/lager"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Task Handlers", func() {
	var (
		logger           lager.Logger
		fakeBBS          *fake_bbs.FakeReceptorBBS
		handler          http.Handler
		responseRecorder *httptest.ResponseRecorder
	)

	BeforeEach(func() {
		fakeBBS = new(fake_bbs.FakeReceptorBBS)
		logger = lager.NewLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
		responseRecorder = httptest.NewRecorder()
	})

	Describe("CreateTask", func() {
		validCreateRequest := receptor.CreateTaskRequest{
			TaskGuid: "task-guid-1",
			Domain:   "test-domain",
			Stack:    "some-stack",
			Actions: []models.ExecutorAction{
				{Action: models.RunAction{Path: "/bin/bash", Args: []string{"echo", "hi"}}},
			},
			MemoryMB:   24,
			DiskMB:     12,
			CpuPercent: 100.0,
			Log:        models.LogConfig{"guid", "source-name"},
			ResultFile: "result-file",
			Annotation: "some annotation",
		}

		expectedTask := models.Task{
			TaskGuid: "task-guid-1",
			Domain:   "test-domain",
			Stack:    "some-stack",
			Actions: []models.ExecutorAction{
				{Action: models.RunAction{Path: "/bin/bash", Args: []string{"echo", "hi"}}},
			},
			MemoryMB:   24,
			DiskMB:     12,
			CpuPercent: 100.0,
			Log:        models.LogConfig{"guid", "source-name"},
			ResultFile: "result-file",
			Annotation: "some annotation",
		}

		BeforeEach(func() {
			handler = NewCreateTaskHandler(fakeBBS, logger)
		})

		Context("when everything succeeds", func() {
			BeforeEach(func(done Done) {
				defer close(done)
				handler.ServeHTTP(responseRecorder, newTestRequest(validCreateRequest))
			})

			It("calls DesireTask on the BBS with the correct task", func() {
				Ω(fakeBBS.DesireTaskCallCount()).Should(Equal(1))
				task := fakeBBS.DesireTaskArgsForCall(0)
				Ω(task).Should(Equal(expectedTask))
			})

			It("responds with 201 CREATED", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusCreated))
			})

			It("responds with an empty body", func() {
				Ω(responseRecorder.Body.String()).Should(Equal(""))
			})
		})

		Context("when the BBS responds with an error", func() {
			BeforeEach(func(done Done) {
				defer close(done)
				fakeBBS.DesireTaskReturns(errors.New("ka-boom"))
				handler.ServeHTTP(responseRecorder, newTestRequest(validCreateRequest))
			})

			It("calls DesireTask on the BBS with the correct task", func() {
				Ω(fakeBBS.DesireTaskCallCount()).Should(Equal(1))
				task := fakeBBS.DesireTaskArgsForCall(0)
				Ω(task.TaskGuid).Should(Equal("task-guid-1"))
			})

			It("responds with 500 INTERNAL ERROR", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusInternalServerError))
			})

			It("responds with a relevant error message", func() {
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.UnknownError,
					Message: "ka-boom",
				})

				Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
			})
		})

		Context("when the requested task is invalid", func() {
			var invalidTask = receptor.CreateTaskRequest{
				TaskGuid: "invalid-task",
			}

			BeforeEach(func(done Done) {
				defer close(done)
				handler.ServeHTTP(responseRecorder, newTestRequest(invalidTask))
			})

			It("does not call DesireTask on the BBS", func() {
				Ω(fakeBBS.DesireTaskCallCount()).Should(Equal(0))
			})

			It("responds with 400 BAD REQUEST", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusBadRequest))
			})

			It("responds with a relevant error message", func() {
				task := models.Task{TaskGuid: "invalid-task"}
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.InvalidTask,
					Message: task.Validate().Error(),
				})
				Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
			})
		})

		Context("when the request does not contain a CreateTaskRequest", func() {
			var garbageRequest = []byte(`hello`)

			BeforeEach(func(done Done) {
				defer close(done)
				handler.ServeHTTP(responseRecorder, newTestRequest(garbageRequest))
			})

			It("does not call DesireTask on the BBS", func() {
				Ω(fakeBBS.DesireTaskCallCount()).Should(Equal(0))
			})

			It("responds with 400 BAD REQUEST", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusBadRequest))
			})

			It("responds with a relevant error message", func() {
				err := json.Unmarshal(garbageRequest, &receptor.CreateTaskRequest{})
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.InvalidJSON,
					Message: err.Error(),
				})
				Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
			})
		})
	})

	Describe("GetAllTasks", func() {
		BeforeEach(func() {
			handler = NewGetAllTasksHandler(fakeBBS, logger)
		})

		Context("when reading tasks from the BBS fails", func() {
			BeforeEach(func() {
				fakeBBS.GetAllTasksReturns([]models.Task{}, errors.New("Something went wrong"))
			})

			It("responds with an error", func() {
				handler.ServeHTTP(responseRecorder, newTestRequest(""))
				Ω(responseRecorder.Code).Should(Equal(http.StatusInternalServerError))
			})
		})

		Context("when reading tasks from BBS succeeds", func() {
			BeforeEach(func() {
				fakeBBS.GetAllTasksReturns([]models.Task{
					{TaskGuid: "task-guid-1", Domain: "domain-1", ContainerHandle: "internal stuff"},
				}, nil)
			})

			It("excludes internal fields", func() {
				handler.ServeHTTP(responseRecorder, newTestRequest(""))
				Ω(responseRecorder.Code).Should(Equal(http.StatusOK))
				Ω(responseRecorder.Body.String()).Should(ContainSubstring("task-guid-1"))
				Ω(responseRecorder.Body.String()).ShouldNot(ContainSubstring("internal stuff"))
			})
		})
	})

	Describe("GetAllTasksByDomain", func() {
		var request *http.Request

		BeforeEach(func() {
			handler = NewGetAllTasksByDomainHandler(fakeBBS, logger)

			var err error
			request, err = http.NewRequest("", "http://example.com?:domain=a-domain", nil)
			Ω(err).ShouldNot(HaveOccurred())
		})

		Context("when reading tasks from the BBS fails", func() {
			BeforeEach(func() {
				fakeBBS.GetAllTasksByDomainReturns([]models.Task{}, errors.New("Something went wrong"))
			})

			It("responds with an error", func() {
				handler.ServeHTTP(responseRecorder, request)
				Ω(responseRecorder.Code).Should(Equal(http.StatusInternalServerError))
			})
		})

		Context("when reading tasks from BBS succeeds", func() {
			BeforeEach(func() {
				fakeBBS.GetAllTasksByDomainReturns([]models.Task{
					{TaskGuid: "task-guid-1", Domain: "domain-1", ContainerHandle: "internal stuff"},
				}, nil)
			})

			It("uses the given domain", func() {
				handler.ServeHTTP(responseRecorder, request)
				Ω(fakeBBS.GetAllTasksByDomainArgsForCall(0)).Should(Equal("a-domain"))
			})

			It("excludes internal fields", func() {
				handler.ServeHTTP(responseRecorder, request)
				Ω(responseRecorder.Code).Should(Equal(http.StatusOK))
				Ω(responseRecorder.Body.String()).Should(ContainSubstring("task-guid-1"))
				Ω(responseRecorder.Body.String()).ShouldNot(ContainSubstring("internal stuff"))
			})
		})
	})
})
