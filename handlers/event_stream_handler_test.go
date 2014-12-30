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

var _ = Describe("Event Stream Handlers", func() {
	var (
		logger           lager.Logger
		fakeBBS          *fake_bbs.FakeReceptorBBS
		responseRecorder *httptest.ResponseRecorder
		handler          *handlers.ActualLRPHandler

		actualLRPPG1I2 = models.ActualLRP{
			ActualLRPKey: models.NewActualLRPKey(
				"process-guid-1",
				2,
				"domain-1",
			),
			ActualLRPContainerKey: models.NewActualLRPContainerKey(
				"instance-guid-1",
				"cell-id-1",
			),
			State: models.ActualLRPStateClaimed,
			Since: 3147,
		}
	)

	BeforeEach(func() {
		fakeBBS = new(fake_bbs.FakeReceptorBBS)
		logger = lager.NewLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
		responseRecorder = httptest.NewRecorder()
		handler = handlers.NewEventStreamHandler(fakeBBS, logger)
	})

	Describe("EventStream", func() {
		Context("when reading LRPs from BBS succeeds", func() {
			BeforeEach(func() {
				actualLRP1 := models.ActualLRP{
					ActualLRPKey: models.NewActualLRPKey(
						"process-guid-0",
						1,
						"domain-0",
					),
					ActualLRPContainerKey: models.NewActualLRPContainerKey(
						"instance-guid-0",
						"cell-id-0",
					),
					State: models.ActualLRPStateRunning,
					Since: 1138,
				}
				actualLRP2 := models.ActualLRP{
					ActualLRPKey: models.NewActualLRPKey(
						"process-guid-1",
						1,
						"domain-1",
					),
					ActualLRPContainerKey: models.NewActualLRPContainerKey(
						"instance-guid-1",
						"cell-id-1",
					),
					State: models.ActualLRPStateRunning,
					Since: 1138,
				}

				fakeBBS.ActualLRPsReturns([]models.ActualLRP{
					actualLRP1,
					actualLRP2,
				}, nil)
				fakeBBS.ActualLRPsByDomainReturns([]models.ActualLRP{actualLRP2}, nil)
			})

			It("call the BBS to retrieve the actual LRPs", func() {
				handler.GetAll(responseRecorder, newTestRequest(""))
				Ω(fakeBBS.ActualLRPsCallCount()).Should(Equal(1))
			})

			It("responds with 200 Status OK", func() {
				handler.GetAll(responseRecorder, newTestRequest(""))
				Ω(responseRecorder.Code).Should(Equal(http.StatusOK))
			})

			Context("when a domain query param is provided", func() {
				It("returns a list of desired lrp responses for the domain", func() {
					request, err := http.NewRequest("", "http://example.com?domain=domain-1", nil)
					Ω(err).ShouldNot(HaveOccurred())

					handler.GetAll(responseRecorder, request)
					response := []receptor.ActualLRPResponse{}
					err = json.Unmarshal(responseRecorder.Body.Bytes(), &response)
					Ω(err).ShouldNot(HaveOccurred())

					Ω(response).Should(HaveLen(1))
					Ω(response[0].Domain).Should(Equal("domain-1"))
				})
			})

			Context("when a domain query param is not provided", func() {
				It("returns a list of desired lrp responses", func() {
					handler.GetAll(responseRecorder, newTestRequest(""))
					response := []receptor.ActualLRPResponse{}
					err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
					Ω(err).ShouldNot(HaveOccurred())

					Ω(response).Should(HaveLen(2))
					Ω(response[0].ProcessGuid).Should(Equal("process-guid-0"))
					Ω(response[1].ProcessGuid).Should(Equal("process-guid-1"))
				})
			})
		})

		Context("when the BBS returns no lrps", func() {
			BeforeEach(func() {
				fakeBBS.ActualLRPsReturns([]models.ActualLRP{}, nil)
			})

			It("call the BBS to retrieve the desired LRP", func() {
				handler.GetAll(responseRecorder, newTestRequest(""))
				Ω(fakeBBS.ActualLRPsCallCount()).Should(Equal(1))
			})

			It("responds with 200 Status OK", func() {
				handler.GetAll(responseRecorder, newTestRequest(""))
				Ω(responseRecorder.Code).Should(Equal(http.StatusOK))
			})

			It("returns an empty list", func() {
				handler.GetAll(responseRecorder, newTestRequest(""))
				Ω(responseRecorder.Body.String()).Should(Equal("[]"))
			})
		})

		Context("when reading from the BBS fails", func() {
			BeforeEach(func() {
				fakeBBS.ActualLRPsReturns([]models.ActualLRP{}, errors.New("Something went wrong"))
			})

			It("responds with an error", func() {
				handler.GetAll(responseRecorder, newTestRequest(""))
				Ω(responseRecorder.Code).Should(Equal(http.StatusInternalServerError))
			})

			It("provides relevant error information", func() {
				handler.GetAll(responseRecorder, newTestRequest(""))
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
