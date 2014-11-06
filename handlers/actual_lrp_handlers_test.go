package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/handlers"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	"github.com/cloudfoundry-incubator/runtime-schema/bbs/fake_bbs"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager"
)

var _ = Describe("Actual LRP Handlers", func() {
	var (
		logger           lager.Logger
		fakeBBS          *fake_bbs.FakeReceptorBBS
		responseRecorder *httptest.ResponseRecorder
		handler          *handlers.ActualLRPHandler
	)

	BeforeEach(func() {
		fakeBBS = new(fake_bbs.FakeReceptorBBS)
		logger = lager.NewLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
		responseRecorder = httptest.NewRecorder()
		handler = handlers.NewActualLRPHandler(fakeBBS, logger)
	})

	Describe("GetAll", func() {
		var actualLRPs = []models.ActualLRP{
			{
				ProcessGuid:  "process-guid-0",
				InstanceGuid: "instance-guid-0",
				ExecutorID:   "executor-id-0",
				Ports: []models.PortMapping{
					{
						ContainerPort: 999,
						HostPort:      888,
					},
				},
			},
			{
				ProcessGuid:  "process-guid-1",
				InstanceGuid: "instance-guid-1",
				ExecutorID:   "executor-id-1",
				Ports: []models.PortMapping{
					{
						ContainerPort: 777,
						HostPort:      666,
					},
				},
			},
		}

		JustBeforeEach(func() {
			handler.GetAll(responseRecorder, newTestRequest(""))
		})

		Context("when reading LRPs from BBS succeeds", func() {
			BeforeEach(func() {
				fakeBBS.GetAllActualLRPsReturns(actualLRPs, nil)
			})

			It("call the BBS to retrieve the actual LRPs", func() {
				Ω(fakeBBS.GetAllActualLRPsCallCount()).Should(Equal(1))
			})

			It("responds with 200 Status OK", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusOK))
			})

			It("returns a list of desired lrp responses", func() {
				response := []receptor.ActualLRPResponse{}
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(response).Should(HaveLen(2))
				for _, actualLRP := range actualLRPs {
					Ω(response).Should(ContainElement(serialization.ActualLRPToResponse(actualLRP)))
				}
			})
		})

		Context("when the BBS returns no lrps", func() {
			BeforeEach(func() {
				fakeBBS.GetAllActualLRPsReturns([]models.ActualLRP{}, nil)
			})

			It("call the BBS to retrieve the desired LRP", func() {
				Ω(fakeBBS.GetAllActualLRPsCallCount()).Should(Equal(1))
			})

			It("responds with 200 Status OK", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusOK))
			})

			It("returns an empty list", func() {
				Ω(responseRecorder.Body.String()).Should(Equal("[]"))
			})
		})

		Context("when reading from the BBS fails", func() {
			BeforeEach(func() {
				fakeBBS.GetAllActualLRPsReturns([]models.ActualLRP{}, errors.New("Something went wrong"))
			})

			It("responds with an error", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusInternalServerError))
			})

			It("provides relevant error information", func() {
				var receptorError receptor.Error
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &receptorError)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(receptorError).Should(Equal(receptor.Error{
					Type:    receptor.UnknownError,
					Message: "Something went wrong",
				}))
			})
		})
	})
})
