package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"

	"github.com/cloudfoundry-incubator/receptor"
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
		validCreateRequest = receptor.CreateTaskRequest{
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
			expectedBody, _ := json.Marshal(receptor.Error{
				Type:    receptor.UnknownError,
				Message: "ka-boom",
			})

			Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
			Ω(responseRecorder.Header().Get("Content-Length")).Should(Equal(strconv.Itoa(len(expectedBody))))
			Ω(responseRecorder.Header().Get("Content-Type")).Should(Equal("application/json"))
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
