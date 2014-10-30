package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/handlers"
	"github.com/cloudfoundry-incubator/runtime-schema/bbs/fake_bbs"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager"
)

var _ = Describe("LRP Handlers", func() {
	var (
		logger           lager.Logger
		fakeBBS          *fake_bbs.FakeReceptorBBS
		responseRecorder *httptest.ResponseRecorder
		handler          *handlers.DesiredLRPHandler
	)

	BeforeEach(func() {
		fakeBBS = new(fake_bbs.FakeReceptorBBS)
		logger = lager.NewLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
		responseRecorder = httptest.NewRecorder()
		handler = handlers.NewDesiredLRPHandler(fakeBBS, logger)
	})

	Describe("Create", func() {
		validCreateLRPRequest := receptor.CreateDesiredLRPRequest{
			ProcessGuid: "the-process-guid",
			Domain:      "the-domain",
			Stack:       "the-stack",
			RootFSPath:  "the-rootfs-path",
			Actions: []models.ExecutorAction{
				{
					Action: models.RunAction{
						Path: "the-path",
					},
				},
			},
		}

		expectedDesiredLRP := models.DesiredLRP{
			ProcessGuid: "the-process-guid",
			Domain:      "the-domain",
			Stack:       "the-stack",
			RootFSPath:  "the-rootfs-path",
			Actions: []models.ExecutorAction{
				{
					Action: models.RunAction{
						Path: "the-path",
					},
				},
			},
		}

		Context("when everything succeeds", func() {
			BeforeEach(func(done Done) {
				defer close(done)
				handler.Create(responseRecorder, newTestRequest(validCreateLRPRequest))
			})

			It("calls DesireLRP on the BBS", func() {
				Ω(fakeBBS.DesireLRPCallCount()).Should(Equal(1))
				desired := fakeBBS.DesireLRPArgsForCall(0)
				Ω(desired).To(Equal(expectedDesiredLRP))
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
				fakeBBS.DesireLRPReturns(errors.New("ka-boom"))
				handler.Create(responseRecorder, newTestRequest(validCreateLRPRequest))
			})

			It("calls DesireLRP on the BBS", func() {
				Ω(fakeBBS.DesireLRPCallCount()).Should(Equal(1))
				desired := fakeBBS.DesireLRPArgsForCall(0)
				Ω(desired).To(Equal(expectedDesiredLRP))
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

		Context("when the desired LRP is invalid", func() {
			var invalidDesiredLRP = receptor.CreateDesiredLRPRequest{}

			BeforeEach(func(done Done) {
				defer close(done)
				handler.Create(responseRecorder, newTestRequest(invalidDesiredLRP))
			})

			It("does not call DesireLRP on the BBS", func() {
				Ω(fakeBBS.DesireLRPCallCount()).Should(Equal(0))
			})

			It("responds with 400 BAD REQUEST", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusBadRequest))
			})

			It("responds with a relevant error message", func() {
				desiredLRP := models.DesiredLRP{}
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.InvalidLRP,
					Message: desiredLRP.Validate().Error(),
				})
				Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
			})
		})

		Context("when the request does not contain a CreateDesiredLRPRequest", func() {
			var garbageRequest = []byte(`farewell`)

			BeforeEach(func(done Done) {
				defer close(done)
				handler.Create(responseRecorder, newTestRequest(garbageRequest))
			})

			It("does not call DesireLRP on the BBS", func() {
				Ω(fakeBBS.DesireLRPCallCount()).Should(Equal(0))
			})

			It("responds with 400 BAD REQUEST", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusBadRequest))
			})

			It("responds with a relevant error message", func() {
				err := json.Unmarshal(garbageRequest, &receptor.CreateDesiredLRPRequest{})
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.InvalidJSON,
					Message: err.Error(),
				})
				Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
			})
		})
	})
})
