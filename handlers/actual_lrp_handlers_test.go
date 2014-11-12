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
				Index:        0,
				InstanceGuid: "instance-guid-0",
				ExecutorID:   "executor-id-0",
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
				Index:        0,
				InstanceGuid: "instance-guid-1",
				ExecutorID:   "executor-id-1",
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
				fakeBBS.GetAllActualLRPsByDomainReturns(actualLRPs[0:1], nil)
			})

			It("call the BBS to retrieve the actual LRPs", func() {
				Ω(fakeBBS.GetAllActualLRPsByDomainCallCount()).Should(Equal(1))
				Ω(fakeBBS.GetAllActualLRPsByDomainArgsForCall(0)).Should(Equal("domain-0"))
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
				fakeBBS.GetAllActualLRPsByDomainReturns([]models.ActualLRP{}, errors.New("Something went wrong"))
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
				fakeBBS.GetAllActualLRPsByDomainReturns([]models.ActualLRP{}, nil)
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
				fakeBBS.GetActualLRPsByProcessGuidReturns(actualLRPs[1:2], nil)
			})

			It("calls the BBS to retrieve the actual LRPs", func() {
				Ω(fakeBBS.GetActualLRPsByProcessGuidCallCount()).Should(Equal(1))
				Ω(fakeBBS.GetActualLRPsByProcessGuidArgsForCall(0)).Should(Equal("process-guid-1"))
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
				fakeBBS.GetActualLRPsByProcessGuidReturns([]models.ActualLRP{}, errors.New("Something went wrong"))
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
				fakeBBS.GetActualLRPsByProcessGuidReturns([]models.ActualLRP{}, nil)
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

		Context("when request includes a valid index query parameter", func() {
			BeforeEach(func() {
				req.Form.Add("index", "0")
			})

			Context("when reading LRPs from BBS succeeds", func() {
				BeforeEach(func() {
					fakeBBS.GetActualLRPsByProcessGuidAndIndexReturns(actualLRPs[1:2], nil)
				})

				It("calls the BBS to retrieve the actual LRPs", func() {
					Ω(fakeBBS.GetActualLRPsByProcessGuidAndIndexCallCount()).Should(Equal(1))
					processGuid, index := fakeBBS.GetActualLRPsByProcessGuidAndIndexArgsForCall(0)
					Ω(processGuid).Should(Equal("process-guid-1"))
					Ω(index).Should(Equal(0))
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
					fakeBBS.GetActualLRPsByProcessGuidAndIndexReturns([]models.ActualLRP{}, errors.New("Something went wrong"))
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
					fakeBBS.GetActualLRPsByProcessGuidAndIndexReturns([]models.ActualLRP{}, nil)
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
		})

		Context("when request includes a bad index query parameter", func() {
			BeforeEach(func() {
				req.Form.Add("index", "not-a-number")
			})

			It("does not call the BBS", func() {
				Ω(fakeBBS.GetActualLRPsByProcessGuidAndIndexCallCount()).Should(Equal(0))
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
})
