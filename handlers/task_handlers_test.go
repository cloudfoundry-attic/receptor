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

var _ = Describe("TaskHandler", func() {
	var (
		logger           lager.Logger
		fakeBBS          *fake_bbs.FakeReceptorBBS
		responseRecorder *httptest.ResponseRecorder
		handler          *TaskHandler
		request          *http.Request
	)

	BeforeEach(func() {
		fakeBBS = new(fake_bbs.FakeReceptorBBS)
		logger = lager.NewLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
		responseRecorder = httptest.NewRecorder()
		handler = NewTaskHandler(fakeBBS, logger)
	})

	Describe("Create", func() {
		validCreateRequest := receptor.TaskCreateRequest{
			TaskGuid:   "task-guid-1",
			Domain:     "test-domain",
			RootFSPath: "docker://docker",
			Stack:      "some-stack",
			Actions: []models.ExecutorAction{
				{Action: models.RunAction{Path: "/bin/bash", Args: []string{"echo", "hi"}}},
			},
			MemoryMB:   24,
			DiskMB:     12,
			CPUWeight:  10,
			Log:        receptor.LogConfig{"guid", "source-name"},
			ResultFile: "result-file",
			Annotation: "some annotation",
		}

		expectedTask := models.Task{
			TaskGuid:   "task-guid-1",
			Domain:     "test-domain",
			RootFSPath: "docker://docker",
			Stack:      "some-stack",
			Actions: []models.ExecutorAction{
				{Action: models.RunAction{Path: "/bin/bash", Args: []string{"echo", "hi"}}},
			},
			MemoryMB:   24,
			DiskMB:     12,
			CPUWeight:  10,
			Log:        models.LogConfig{"guid", "source-name"},
			ResultFile: "result-file",
			Annotation: "some annotation",
		}

		Context("when everything succeeds", func() {
			JustBeforeEach(func(done Done) {
				defer close(done)
				handler.Create(responseRecorder, newTestRequest(validCreateRequest))
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

			Context("when env vars are specified", func() {
				BeforeEach(func() {
					validCreateRequest.EnvironmentVariables = []receptor.EnvironmentVariable{
						{Name: "var1", Value: "val1"},
						{Name: "var2", Value: "val2"},
					}
				})

				AfterEach(func() {
					validCreateRequest.EnvironmentVariables = []receptor.EnvironmentVariable{}
				})

				It("passes them to the BBS", func() {
					Ω(fakeBBS.DesireTaskCallCount()).Should(Equal(1))
					task := fakeBBS.DesireTaskArgsForCall(0)
					Ω(task.EnvironmentVariables).Should(Equal([]models.EnvironmentVariable{
						{Name: "var1", Value: "val1"},
						{Name: "var2", Value: "val2"},
					}))
				})
			})

			Context("when no env vars are specified", func() {
				It("passes a nil slice to the BBS", func() {
					Ω(fakeBBS.DesireTaskCallCount()).Should(Equal(1))
					task := fakeBBS.DesireTaskArgsForCall(0)
					Ω(task.EnvironmentVariables).Should(BeNil())
				})
			})
		})

		Context("when the BBS responds with an error", func() {
			BeforeEach(func(done Done) {
				defer close(done)
				fakeBBS.DesireTaskReturns(errors.New("ka-boom"))
				handler.Create(responseRecorder, newTestRequest(validCreateRequest))
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
			var invalidTask = receptor.TaskCreateRequest{
				TaskGuid: "invalid-task",
			}

			BeforeEach(func(done Done) {
				defer close(done)
				handler.Create(responseRecorder, newTestRequest(invalidTask))
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

		Context("when the request does not contain a TaskCreateRequest", func() {
			var garbageRequest = []byte(`hello`)

			BeforeEach(func(done Done) {
				defer close(done)
				handler.Create(responseRecorder, newTestRequest(garbageRequest))
			})

			It("does not call DesireTask on the BBS", func() {
				Ω(fakeBBS.DesireTaskCallCount()).Should(Equal(0))
			})

			It("responds with 400 BAD REQUEST", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusBadRequest))
			})

			It("responds with a relevant error message", func() {
				err := json.Unmarshal(garbageRequest, &receptor.TaskCreateRequest{})
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.InvalidJSON,
					Message: err.Error(),
				})
				Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
			})
		})
	})

	Describe("GetAll", func() {
		Context("when reading tasks from the BBS fails", func() {
			BeforeEach(func() {
				fakeBBS.GetAllTasksReturns([]models.Task{}, errors.New("Something went wrong"))
			})

			It("responds with an error", func() {
				handler.GetAll(responseRecorder, newTestRequest(""))
				Ω(responseRecorder.Code).Should(Equal(http.StatusInternalServerError))
			})
		})

		Context("when reading tasks from BBS succeeds", func() {
			BeforeEach(func() {
				fakeBBS.GetAllTasksReturns([]models.Task{
					{TaskGuid: "task-guid-1", Domain: "domain-1"},
				}, nil)
			})

			It("excludes internal fields", func() {
				handler.GetAll(responseRecorder, newTestRequest(""))
				Ω(responseRecorder.Code).Should(Equal(http.StatusOK))
				Ω(responseRecorder.Body.String()).Should(ContainSubstring("task-guid-1"))
				Ω(responseRecorder.Body.String()).ShouldNot(ContainSubstring("internal stuff"))
			})
		})
	})

	Describe("GetAllByDomain", func() {
		BeforeEach(func() {
			var err error
			request, err = http.NewRequest("", "http://example.com?:domain=a-domain", nil)
			Ω(err).ShouldNot(HaveOccurred())
		})

		Context("when reading tasks from the BBS fails", func() {
			BeforeEach(func() {
				fakeBBS.GetAllTasksByDomainReturns([]models.Task{}, errors.New("Something went wrong"))
			})

			It("responds with an error", func() {
				handler.GetAllByDomain(responseRecorder, request)
				Ω(responseRecorder.Code).Should(Equal(http.StatusInternalServerError))
			})
		})

		Context("when reading tasks from BBS succeeds", func() {
			BeforeEach(func() {
				fakeBBS.GetAllTasksByDomainReturns([]models.Task{
					{TaskGuid: "task-guid-1", Domain: "domain-1"},
				}, nil)
			})

			It("uses the given domain", func() {
				handler.GetAllByDomain(responseRecorder, request)
				Ω(fakeBBS.GetAllTasksByDomainArgsForCall(0)).Should(Equal("a-domain"))
			})

			It("excludes internal fields", func() {
				handler.GetAllByDomain(responseRecorder, request)
				Ω(responseRecorder.Code).Should(Equal(http.StatusOK))
				Ω(responseRecorder.Body.String()).Should(ContainSubstring("task-guid-1"))
				Ω(responseRecorder.Body.String()).ShouldNot(ContainSubstring("internal stuff"))
			})
		})
	})

	Describe("GetByGuid", func() {
		BeforeEach(func() {
			var err error
			request, err = http.NewRequest("", "http://example.com?:task_guid=the-task-guid", nil)
			Ω(err).ShouldNot(HaveOccurred())
		})

		Context("when reading the task from the BBS fails", func() {
			BeforeEach(func() {
				fakeBBS.GetTaskByGuidReturns(models.Task{}, errors.New("Something went wrong"))
			})

			It("responds with an error", func() {
				handler.GetByGuid(responseRecorder, request)
				Ω(responseRecorder.Code).Should(Equal(http.StatusInternalServerError))
			})
		})

		Context("when the task is successfully found in the BBS", func() {
			BeforeEach(func() {
				fakeBBS.GetTaskByGuidReturns(models.Task{
					TaskGuid: "task-guid-1",
					Domain:   "domain-1",
				}, nil)
			})

			It("retrieves the task by the given guid", func() {
				handler.GetByGuid(responseRecorder, request)
				Ω(fakeBBS.GetTaskByGuidArgsForCall(0)).Should(Equal("the-task-guid"))
			})

			It("excludes internal fields", func() {
				handler.GetByGuid(responseRecorder, request)
				Ω(responseRecorder.Code).Should(Equal(http.StatusOK))
				Ω(responseRecorder.Body.String()).Should(ContainSubstring("task-guid-1"))
				Ω(responseRecorder.Body.String()).ShouldNot(ContainSubstring("internal stuff"))
			})
		})
	})

	Describe("Delete", func() {
		Context("when marking the task as resolving fails", func() {
			BeforeEach(func() {
				var err error
				request, err = http.NewRequest("", "http://example.com?:task_guid=the-task-guid", nil)
				Ω(err).ShouldNot(HaveOccurred())
				fakeBBS.ResolvingTaskReturns(errors.New("Failed to resolve task"))
			})

			It("responds with an error", func() {
				handler.Delete(responseRecorder, request)
				Ω(responseRecorder.Code).Should(Equal(http.StatusInternalServerError))
			})

			It("does not try to resolve the task", func() {
				handler.Delete(responseRecorder, request)
				Ω(fakeBBS.ResolveTaskCallCount()).Should(BeZero())
			})
		})

		Context("when task cannot be resolved", func() {
			BeforeEach(func() {
				var err error
				request, err = http.NewRequest("", "http://example.com?:task_guid=the-task-guid", nil)
				Ω(err).ShouldNot(HaveOccurred())
				fakeBBS.ResolveTaskReturns(errors.New("Failed to resolve task"))
			})

			It("responds with an error", func() {
				handler.Delete(responseRecorder, request)
				Ω(responseRecorder.Code).Should(Equal(http.StatusInternalServerError))
			})
		})
	})
})
