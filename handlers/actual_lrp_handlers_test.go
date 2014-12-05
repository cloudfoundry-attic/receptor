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

		actualLRPs = []models.ActualLRP{
			{
				ProcessGuid:  "process-guid-0",
				Index:        1,
				InstanceGuid: "instance-guid-0",
				CellID:       "cell-id-0",
				Domain:       "domain-0",
				Ports: []models.PortMapping{
					{
						ContainerPort: 999,
						HostPort:      888,
					},
				},
			},
			{
				ProcessGuid:  "process-guid-1",
				Index:        2,
				InstanceGuid: "instance-guid-1",
				CellID:       "cell-id-1",
				Domain:       "domain-1",
				Ports: []models.PortMapping{
					{
						ContainerPort: 777,
						HostPort:      666,
					},
				},
			},
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
		JustBeforeEach(func() {
			handler.GetAll(responseRecorder, newTestRequest(""))
		})

		Context("when reading LRPs from BBS succeeds", func() {
			BeforeEach(func() {
				fakeBBS.ActualLRPsReturns(actualLRPs, nil)
			})

			It("call the BBS to retrieve the actual LRPs", func() {
				Ω(fakeBBS.ActualLRPsCallCount()).Should(Equal(1))
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
				fakeBBS.ActualLRPsReturns([]models.ActualLRP{}, nil)
			})

			It("call the BBS to retrieve the desired LRP", func() {
				Ω(fakeBBS.ActualLRPsCallCount()).Should(Equal(1))
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
				fakeBBS.ActualLRPsReturns([]models.ActualLRP{}, errors.New("Something went wrong"))
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

	Describe("GetAllByDomain", func() {
		var req *http.Request

		BeforeEach(func() {
			req = newTestRequest("")
			req.Form = url.Values{":domain": []string{"domain-0"}}
		})

		JustBeforeEach(func() {
			handler.GetAllByDomain(responseRecorder, req)
		})

		Context("when reading LRPs from BBS succeeds", func() {
			BeforeEach(func() {
				fakeBBS.ActualLRPsByDomainReturns(actualLRPs[0:1], nil)
			})

			It("call the BBS to retrieve the actual LRPs", func() {
				Ω(fakeBBS.ActualLRPsByDomainCallCount()).Should(Equal(1))
				Ω(fakeBBS.ActualLRPsByDomainArgsForCall(0)).Should(Equal("domain-0"))
			})

			It("responds with 200 Status OK", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusOK))
			})

			It("returns a list of actual lrp responses", func() {
				response := []receptor.ActualLRPResponse{}
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(response).Should(HaveLen(1))
				Ω(response).Should(ContainElement(serialization.ActualLRPToResponse(actualLRPs[0])))
			})
		})

		Context("when reading LRPs from BBS fails", func() {
			BeforeEach(func() {
				fakeBBS.ActualLRPsByDomainReturns([]models.ActualLRP{}, errors.New("Something went wrong"))
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

		Context("when the BBS doesn't return any actual LRPs", func() {
			BeforeEach(func() {
				fakeBBS.ActualLRPsByDomainReturns([]models.ActualLRP{}, nil)
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

		Context("when the request does not contain a domain parameter", func() {
			BeforeEach(func() {
				req.Form = url.Values{}
			})

			It("responds with 400 Bad Request", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusBadRequest))
			})

			It("responds with a relevant error message", func() {
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.InvalidRequest,
					Message: "domain missing from request",
				})

				Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
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
				fakeBBS.ActualLRPsByProcessGuidReturns(actualLRPs[1:2], nil)
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
				Ω(response).Should(ContainElement(serialization.ActualLRPToResponse(actualLRPs[1])))
			})
		})

		Context("when reading LRPs from BBS fails", func() {
			BeforeEach(func() {
				fakeBBS.ActualLRPsByProcessGuidReturns([]models.ActualLRP{}, errors.New("Something went wrong"))
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
				fakeBBS.ActualLRPsByProcessGuidReturns([]models.ActualLRP{}, nil)
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

	Describe("GetAllByProcessGuidAndIndex", func() {
		var req *http.Request

		BeforeEach(func() {
			req = newTestRequest("")
			req.Form = url.Values{":process_guid": []string{"process-guid-1"},
				":index": []string{"1"},
			}
		})

		JustBeforeEach(func() {
			handler.GetByProcessGuidAndIndex(responseRecorder, req)
		})

		Context("when reading LRPs from BBS succeeds", func() {
			BeforeEach(func() {
				fakeBBS.ActualLRPByProcessGuidAndIndexReturns(&actualLRPs[1], nil)
			})

			It("calls the BBS to retrieve the actual LRPs", func() {
				Ω(fakeBBS.ActualLRPByProcessGuidAndIndexCallCount()).Should(Equal(1))
				processGuid, index := fakeBBS.ActualLRPByProcessGuidAndIndexArgsForCall(0)
				Ω(processGuid).Should(Equal("process-guid-1"))
				Ω(index).Should(Equal(1))
			})

			It("responds with 200 Status OK", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusOK))
			})

			It("returns an actual lrp response", func() {
				response := receptor.ActualLRPResponse{}
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(response).Should(Equal(serialization.ActualLRPToResponse(actualLRPs[1])))
			})
		})

		Context("when reading LRPs from BBS fails", func() {
			BeforeEach(func() {
				fakeBBS.ActualLRPByProcessGuidAndIndexReturns(nil, errors.New("Something went wrong"))
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
				fakeBBS.ActualLRPByProcessGuidAndIndexReturns(nil, nil)
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
					fakeBBS.ActualLRPByProcessGuidAndIndexReturns(&actualLRPs[1], nil)
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
					fakeBBS.ActualLRPByProcessGuidAndIndexReturns(nil, nil)
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
					fakeBBS.ActualLRPByProcessGuidAndIndexReturns(nil, errors.New("Something went wrong"))
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
					fakeBBS.ActualLRPByProcessGuidAndIndexReturns(&actualLRPs[1], nil)
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
