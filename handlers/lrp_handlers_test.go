package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"

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
		validCreateLRPRequest := receptor.DesiredLRPCreateRequest{
			ProcessGuid: "the-process-guid",
			Domain:      "the-domain",
			Stack:       "the-stack",
			RootFSPath:  "the-rootfs-path",
			Instances:   1,
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
			Instances:   1,
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
			var invalidDesiredLRP = receptor.DesiredLRPCreateRequest{}

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

		Context("when the request does not contain a DesiredLRPCreateRequest", func() {
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
				err := json.Unmarshal(garbageRequest, &receptor.DesiredLRPCreateRequest{})
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.InvalidJSON,
					Message: err.Error(),
				})
				Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
			})
		})
	})

	Describe("Update", func() {
		expectedProcessGuid := "some-guid"
		instances := 15
		annotation := "new-annotation"
		routes := []string{"new-route-1", "new-route-2"}

		validUpdateRequest := receptor.DesiredLRPUpdateRequest{
			Instances:  &instances,
			Annotation: &annotation,
			Routes:     routes,
		}

		expectedUpdate := models.DesiredLRPUpdate{
			Instances:  &instances,
			Annotation: &annotation,
			Routes:     routes,
		}

		var req *http.Request

		BeforeEach(func() {
			req = newTestRequest(validUpdateRequest)
			req.Form = url.Values{":process_guid": []string{expectedProcessGuid}}
		})

		Context("when everything succeeds", func() {
			BeforeEach(func(done Done) {
				defer close(done)
				handler.Update(responseRecorder, req)
			})

			It("calls UpdateDesiredLRP on the BBS", func() {
				Ω(fakeBBS.UpdateDesiredLRPCallCount()).Should(Equal(1))
				processGuid, update := fakeBBS.UpdateDesiredLRPArgsForCall(0)
				Ω(processGuid).Should(Equal(expectedProcessGuid))
				Ω(update).Should(Equal(expectedUpdate))
			})

			It("responds with 204 NO CONTENT", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusNoContent))
			})

			It("responds with an empty body", func() {
				Ω(responseRecorder.Body.String()).Should(Equal(""))
			})
		})

		Context("when the :process_guid is blank", func() {
			BeforeEach(func() {
				req = newTestRequest(validUpdateRequest)
				handler.Update(responseRecorder, req)
			})

			It("responds with 400 BAD REQUEST", func() {
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

		Context("when the BBS responds with an error", func() {
			BeforeEach(func(done Done) {
				defer close(done)
				fakeBBS.UpdateDesiredLRPReturns(errors.New("ka-boom"))
				handler.Update(responseRecorder, req)
			})

			It("calls UpdateDesiredLRP on the BBS", func() {
				Ω(fakeBBS.UpdateDesiredLRPCallCount()).Should(Equal(1))
				processGuid, update := fakeBBS.UpdateDesiredLRPArgsForCall(0)
				Ω(processGuid).Should(Equal(expectedProcessGuid))
				Ω(update).Should(Equal(expectedUpdate))
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

		Context("when the request does not contain an DesiredLRPUpdateRequest", func() {
			var garbageRequest = []byte(`farewell`)

			BeforeEach(func(done Done) {
				defer close(done)
				req = newTestRequest(garbageRequest)
				req.Form = url.Values{":process_guid": []string{expectedProcessGuid}}
				handler.Update(responseRecorder, req)
			})

			It("does not call DesireLRP on the BBS", func() {
				Ω(fakeBBS.UpdateDesiredLRPCallCount()).Should(Equal(0))
			})

			It("responds with 400 BAD REQUEST", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusBadRequest))
			})

			It("responds with a relevant error message", func() {
				err := json.Unmarshal(garbageRequest, &receptor.DesiredLRPUpdateRequest{})
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.InvalidJSON,
					Message: err.Error(),
				})
				Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
			})
		})
	})

	Describe("GetAll", func() {
		JustBeforeEach(func() {
			handler.GetAll(responseRecorder, newTestRequest(""))
		})

		Context("when reading tasks from BBS succeeds", func() {
			BeforeEach(func() {
				fakeBBS.GetAllDesiredLRPsReturns([]models.DesiredLRP{
					{ProcessGuid: "process-guid-0", Domain: "domain-1"},
					{ProcessGuid: "process-guid-1", Domain: "domain-1"},
				}, nil)
			})

			It("responds with 200 Status OK", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusOK))
			})

			It("returns a list of desired lrp responses", func() {
				response := []receptor.DesiredLRPResponse{}
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(response).Should(HaveLen(2))
				Ω(response[0].ProcessGuid).Should(Equal("process-guid-0"))
				Ω(response[1].ProcessGuid).Should(Equal("process-guid-1"))
			})
		})

		Context("when the BBS returns no lrps", func() {
			BeforeEach(func() {
				fakeBBS.GetAllDesiredLRPsReturns([]models.DesiredLRP{}, nil)
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
				fakeBBS.GetAllDesiredLRPsReturns([]models.DesiredLRP{}, errors.New("Something went wrong"))
			})

			It("responds with an error", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusInternalServerError))
			})
		})
	})

	Describe("GetAllByDomain", func() {
		var req *http.Request

		BeforeEach(func() {
			req = newTestRequest("")
			req.Form = url.Values{":domain": []string{"domain-1"}}
		})

		JustBeforeEach(func() {
			handler.GetAllByDomain(responseRecorder, req)
		})

		Context("when reading LRPs by domain from BBS succeeds", func() {
			BeforeEach(func() {
				fakeBBS.GetAllDesiredLRPsByDomainReturns([]models.DesiredLRP{
					{ProcessGuid: "process-guid-0", Domain: "domain-1"},
					{ProcessGuid: "process-guid-1", Domain: "domain-1"},
				}, nil)
			})

			It("responds with 200 Status OK", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusOK))
			})

			It("returns a list of desired lrp responses", func() {
				response := []receptor.DesiredLRPResponse{}
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(response).Should(HaveLen(2))
				Ω(response[0].ProcessGuid).Should(Equal("process-guid-0"))
				Ω(response[1].ProcessGuid).Should(Equal("process-guid-1"))
			})
		})

		Context("when the BBS returns no lrps", func() {
			BeforeEach(func() {
				fakeBBS.GetAllDesiredLRPsByDomainReturns([]models.DesiredLRP{}, nil)
			})

			It("responds with 200 Status OK", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusOK))
			})

			It("returns an empty list", func() {
				Ω(responseRecorder.Body.String()).Should(Equal("[]"))
			})
		})

		Context("when the :domain is blank", func() {
			BeforeEach(func() {
				req.Form = url.Values{}
			})

			It("responds with 400 BAD REQUEST", func() {
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

		Context("when reading from the BBS fails", func() {
			BeforeEach(func() {
				fakeBBS.GetAllDesiredLRPsByDomainReturns([]models.DesiredLRP{}, errors.New("Something went wrong"))
			})

			It("responds with an error", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusInternalServerError))
			})
		})
	})
})
