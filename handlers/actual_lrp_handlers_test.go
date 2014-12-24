package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/handlers"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	"github.com/cloudfoundry-incubator/runtime-schema/bbs/bbserrors"
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
		handler = handlers.NewActualLRPHandler(fakeBBS, logger)
	})

	Describe("GetAll", func() {
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

	Describe("GetAllByProcessGuid", func() {
		var req *http.Request

		BeforeEach(func() {
			req = newTestRequest("")
			req.Form = url.Values{":process_guid": []string{"process-guid-1"}}
		})

		JustBeforeEach(func() {
			handler.GetAllByProcessGuid(responseRecorder, req)
		})

		Context("when reading LRPs from BBS succeeds", func() {
			BeforeEach(func() {
				fakeBBS.ActualLRPsByProcessGuidReturns(models.ActualLRPsByIndex{
					2: actualLRPPG1I2,
				}, nil)
			})

			It("calls the BBS to retrieve the actual LRPs", func() {
				Ω(fakeBBS.ActualLRPsByProcessGuidCallCount()).Should(Equal(1))
				Ω(fakeBBS.ActualLRPsByProcessGuidArgsForCall(0)).Should(Equal("process-guid-1"))
			})

			It("responds with 200 Status OK", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusOK))
			})

			It("returns a list of actual lrp responses", func() {
				response := []receptor.ActualLRPResponse{}
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(response).Should(HaveLen(1))
				Ω(response).Should(ContainElement(serialization.ActualLRPToResponse(actualLRPPG1I2)))
			})
		})

		Context("when reading LRPs from BBS fails", func() {
			BeforeEach(func() {
				fakeBBS.ActualLRPsByProcessGuidReturns(nil, errors.New("Something went wrong"))
			})

			It("responds with a 500 Internal Error", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusInternalServerError))
			})

			It("responds with a relevant error message", func() {
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.UnknownError,
					Message: "Something went wrong",
				})

				Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
			})
		})

		Context("when the BBS does not return any actual LRPs", func() {
			BeforeEach(func() {
				fakeBBS.ActualLRPsByProcessGuidReturns(models.ActualLRPsByIndex{}, nil)
			})

			It("responds with 200 Status OK", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusOK))
			})

			It("returns an empty list", func() {
				response := []receptor.ActualLRPResponse{}
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(response).Should(HaveLen(0))
			})
		})

		Context("when the request does not contain a process_guid parameter", func() {
			BeforeEach(func() {
				req.Form = url.Values{}
			})

			It("responds with 400 Bad Request", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusBadRequest))
			})

			It("responds with a relevant error message", func() {
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.InvalidRequest,
					Message: "process_guid missing from request",
				})

				Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
			})
		})

	})

	Describe("GetByProcessGuidAndIndex", func() {
		var req *http.Request

		BeforeEach(func() {
			req = newTestRequest("")
			req.Form = url.Values{":process_guid": []string{"process-guid-1"},
				":index": []string{"2"},
			}
		})

		JustBeforeEach(func() {
			handler.GetByProcessGuidAndIndex(responseRecorder, req)
		})

		Context("when reading LRPs from BBS succeeds", func() {
			BeforeEach(func() {
				fakeBBS.ActualLRPByProcessGuidAndIndexReturns(actualLRPPG1I2, nil)
			})

			It("calls the BBS to retrieve the actual LRPs", func() {
				Ω(fakeBBS.ActualLRPByProcessGuidAndIndexCallCount()).Should(Equal(1))
				processGuid, index := fakeBBS.ActualLRPByProcessGuidAndIndexArgsForCall(0)
				Ω(processGuid).Should(Equal("process-guid-1"))
				Ω(index).Should(Equal(2))
			})

			It("responds with 200 Status OK", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusOK))
			})

			It("returns an actual lrp response", func() {
				response := receptor.ActualLRPResponse{}
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(response).Should(Equal(serialization.ActualLRPToResponse(actualLRPPG1I2)))
			})
		})

		Context("when reading LRPs from BBS fails", func() {
			BeforeEach(func() {
				fakeBBS.ActualLRPByProcessGuidAndIndexReturns(models.ActualLRP{}, errors.New("Something went wrong"))
			})

			It("responds with a 500 Internal Error", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusInternalServerError))
			})

			It("responds with a relevant error message", func() {
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.UnknownError,
					Message: "Something went wrong",
				})

				Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
			})
		})

		Context("when the BBS does not return any actual LRP", func() {
			BeforeEach(func() {
				fakeBBS.ActualLRPByProcessGuidAndIndexReturns(models.ActualLRP{}, bbserrors.ErrStoreResourceNotFound)
			})

			It("responds with 404 Not Found", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusNotFound))
			})
		})

		Context("when request includes a bad index query parameter", func() {
			BeforeEach(func() {
				req.Form.Set(":index", "not-a-number")
			})

			It("does not call the BBS", func() {
				Ω(fakeBBS.ActualLRPByProcessGuidAndIndexCallCount()).Should(Equal(0))
			})

			It("responds with 400 Bad Request", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusBadRequest))
			})

			It("responds with a relevant error message", func() {
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.InvalidRequest,
					Message: "index not a number",
				})

				Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
			})
		})
	})

	Describe("KillByProcessGuidAndIndex", func() {
		var req *http.Request

		BeforeEach(func() {
			req = newTestRequest("")
			req.Form = url.Values{":process_guid": []string{"process-guid-1"}}
		})

		JustBeforeEach(func() {
			handler.KillByProcessGuidAndIndex(responseRecorder, req)
		})

		Context("when request includes a valid index query parameter", func() {
			BeforeEach(func() {
				req.Form.Add(":index", "0")
			})

			Context("when reading LRPs from BBS succeeds", func() {
				BeforeEach(func() {
					fakeBBS.ActualLRPByProcessGuidAndIndexReturns(actualLRPPG1I2, nil)
					fakeBBS.RequestStopLRPInstanceReturns(nil)
				})

				It("calls the BBS to retrieve the actual LRPs", func() {
					Ω(fakeBBS.ActualLRPByProcessGuidAndIndexCallCount()).Should(Equal(1))
					processGuid, index := fakeBBS.ActualLRPByProcessGuidAndIndexArgsForCall(0)
					Ω(processGuid).Should(Equal("process-guid-1"))
					Ω(index).Should(Equal(0))
				})

				It("calls the BBS to request stop LRP instances", func() {
					Ω(fakeBBS.RequestStopLRPInstanceCallCount()).Should(Equal(1))
					stopLRPInstance := fakeBBS.RequestStopLRPInstanceArgsForCall(0)
					Ω(stopLRPInstance.ProcessGuid).Should(Equal("process-guid-1"))
					Ω(stopLRPInstance.Index).Should(Equal(2))
				})

				It("responds with 204 Status NO CONTENT", func() {
					Ω(responseRecorder.Code).Should(Equal(http.StatusNoContent))
				})
			})

			Context("when the BBS returns no lrps", func() {
				BeforeEach(func() {
					fakeBBS.ActualLRPByProcessGuidAndIndexReturns(models.ActualLRP{}, bbserrors.ErrStoreResourceNotFound)
				})

				It("call the BBS to retrieve the desired LRP", func() {
					Ω(fakeBBS.ActualLRPByProcessGuidAndIndexCallCount()).Should(Equal(1))
				})

				It("responds with 404 Status NOT FOUND", func() {
					Ω(responseRecorder.Code).Should(Equal(http.StatusNotFound))
				})
			})

			Context("when reading LRPs from BBS fails", func() {
				BeforeEach(func() {
					fakeBBS.ActualLRPByProcessGuidAndIndexReturns(models.ActualLRP{}, errors.New("Something went wrong"))
				})

				It("does not call the BBS to request stopping instances", func() {
					Ω(fakeBBS.RequestStopLRPInstanceCallCount()).Should(Equal(0))
				})

				It("responds with a 500 Internal Error", func() {
					Ω(responseRecorder.Code).Should(Equal(http.StatusInternalServerError))
				})

				It("responds with a relevant error message", func() {
					expectedBody, _ := json.Marshal(receptor.Error{
						Type:    receptor.UnknownError,
						Message: "Something went wrong",
					})

					Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
				})
			})

			Context("when stopping instances on the BBS fails", func() {
				BeforeEach(func() {
					fakeBBS.ActualLRPByProcessGuidAndIndexReturns(actualLRPPG1I2, nil)
					fakeBBS.RequestStopLRPInstanceReturns(errors.New("Something went wrong"))
				})

				It("responds with a 500 Internal Error", func() {
					Ω(responseRecorder.Code).Should(Equal(http.StatusInternalServerError))
				})

				It("responds with a relevant error message", func() {
					expectedBody, _ := json.Marshal(receptor.Error{
						Type:    receptor.UnknownError,
						Message: "Something went wrong",
					})

					Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
				})
			})
		})

		Context("when the index is not specified", func() {
			It("does not call the BBS at all", func() {
				Ω(fakeBBS.ActualLRPByProcessGuidAndIndexCallCount()).Should(Equal(0))
				Ω(fakeBBS.RequestStopLRPInstanceCallCount()).Should(Equal(0))
			})

			It("responds with 400 Bad Request", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusBadRequest))
			})

			It("responds with a relevant error message", func() {
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.InvalidRequest,
					Message: "index missing from request",
				})

				Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
			})
		})

		Context("when the index is not a number", func() {
			BeforeEach(func() {
				req.Form.Add(":index", "not-a-number")
			})

			It("does not call the BBS at all", func() {
				Ω(fakeBBS.ActualLRPByProcessGuidAndIndexCallCount()).Should(Equal(0))
				Ω(fakeBBS.RequestStopLRPInstanceCallCount()).Should(Equal(0))
			})

			It("responds with 400 Bad Request", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusBadRequest))
			})

			It("responds with a relevant error message", func() {
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.InvalidRequest,
					Message: "index not a number",
				})

				Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
			})
		})

		Context("when the process guid is not specified", func() {
			BeforeEach(func() {
				req.Form = url.Values{}
			})

			It("does not call the BBS at all", func() {
				Ω(fakeBBS.ActualLRPByProcessGuidAndIndexCallCount()).Should(Equal(0))
				Ω(fakeBBS.RequestStopLRPInstanceCallCount()).Should(Equal(0))
			})

			It("responds with 400 Bad Request", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusBadRequest))
			})

			It("responds with a relevant error message", func() {
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.InvalidRequest,
					Message: "process_guid missing from request",
				})

				Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
			})
		})
	})
})
