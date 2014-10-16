package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/cloudfoundry-incubator/receptor/api"
	. "github.com/cloudfoundry-incubator/receptor/handlers"
	"github.com/cloudfoundry-incubator/runtime-schema/bbs/fake_bbs"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/pivotal-golang/lager"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Create Task Handler", func() {
	var (
		fakeBBS            *fake_bbs.FakeReceptorBBS
		handler            http.Handler
		responseRecorder   *httptest.ResponseRecorder
		validCreateRequest = api.CreateTaskRequest{
			TaskGuid: "task-guid-1",
			Domain:   "test-domain",
			Stack:    "some-stack",
			Actions: []models.ExecutorAction{
				{Action: models.RunAction{Path: "/bin/bash", Args: []string{"echo", "hi"}}},
			},
		}
	)

	BeforeEach(func() {
		fakeBBS = new(fake_bbs.FakeReceptorBBS)
		logger := lager.NewLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
		handler = NewCreateTaskHandler(fakeBBS, logger)
		responseRecorder = httptest.NewRecorder()
	})

	Context("when everything succeeds", func() {
		BeforeEach(func(done Done) {
			defer close(done)
			handler.ServeHTTP(responseRecorder, newTestRequest(validCreateRequest))
		})

		It("calls DesireTask on the BBS with the correct task", func() {
			Ω(fakeBBS.DesireTaskCallCount()).Should(Equal(1))
			task := fakeBBS.DesireTaskArgsForCall(0)
			Ω(task.TaskGuid).Should(Equal("task-guid-1"))
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
			expectedError := api.ErrorResponse{
				Error: "ka-boom",
			}
			expectedBody := expectedError.JSONReader()
			Ω(responseRecorder.Body.String()).Should(Equal(expectedBody.String()))
		})
	})

	Context("when the requested task is invalid", func() {
		var invalidTask = api.CreateTaskRequest{
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
			_, err := invalidTask.ToTask()
			expectedError := api.ErrorResponse{
				Error: err.Error(),
			}
			expectedBody := expectedError.JSONReader()
			Ω(responseRecorder.Body.String()).Should(Equal(expectedBody.String()))
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
			err := json.Unmarshal(garbageRequest, &api.CreateTaskRequest{})
			expectedError := api.ErrorResponse{
				Error: err.Error(),
			}
			expectedBody := expectedError.JSONReader()
			Ω(responseRecorder.Body.String()).Should(Equal(expectedBody.String()))
		})
	})

})
