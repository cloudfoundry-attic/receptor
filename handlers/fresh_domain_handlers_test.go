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

var _ = Describe("Fresh Domain Handlers", func() {
	var (
		logger           lager.Logger
		fakeBBS          *fake_bbs.FakeReceptorBBS
		responseRecorder *httptest.ResponseRecorder
		handler          *handlers.FreshDomainHandler
	)

	BeforeEach(func() {
		fakeBBS = new(fake_bbs.FakeReceptorBBS)
		logger = lager.NewLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
		responseRecorder = httptest.NewRecorder()
		handler = handlers.NewFreshDomainHandler(fakeBBS, logger)
	})

	Describe("Create", func() {
		Context("with a structured request", func() {
			var freshDomainCreateRequest receptor.FreshDomainCreateRequest
			var expectedFreshness models.Freshness

			BeforeEach(func() {
				freshDomainCreateRequest = receptor.FreshDomainCreateRequest{
					Domain:       "domain-1",
					TTLInSeconds: 1000,
				}

				expectedFreshness = models.Freshness{
					Domain:       "domain-1",
					TTLInSeconds: 1000,
				}
			})

			JustBeforeEach(func() {
				handler.Create(responseRecorder, newTestRequest(freshDomainCreateRequest))
			})

			Context("when the call to the BBS succeeds", func() {
				It("calls BumpFreshness on the BBS", func() {
					Ω(fakeBBS.BumpFreshnessCallCount()).Should(Equal(1))
					freshness := fakeBBS.BumpFreshnessArgsForCall(0)
					Ω(freshness).To(Equal(expectedFreshness))
				})

				It("responds with 204 Status NO CONTENT", func() {
					Ω(responseRecorder.Code).Should(Equal(http.StatusNoContent))
				})

				It("responds with an empty body", func() {
					Ω(responseRecorder.Body.String()).Should(Equal(""))
				})
			})

			Context("when the call to the BBS fails", func() {
				BeforeEach(func() {
					fakeBBS.BumpFreshnessReturns(errors.New("ka-boom"))
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

			Context("when the request corresponds to an invalid freshness", func() {
				BeforeEach(func() {
					freshDomainCreateRequest = receptor.FreshDomainCreateRequest{
						Domain:       "",
						TTLInSeconds: -1000,
					}

					expectedFreshness = models.Freshness{
						Domain:       "",
						TTLInSeconds: -1000,
					}
				})

				It("does not call BumpFreshness on the BBS", func() {
					Ω(fakeBBS.BumpFreshnessCallCount()).Should(Equal(0))
				})

				It("responds with 400 BAD REQUEST", func() {
					Ω(responseRecorder.Code).Should(Equal(http.StatusBadRequest))
				})

				It("responds with a relevant error message", func() {
					expectedBody, _ := json.Marshal(receptor.Error{
						Type:    receptor.InvalidFreshness,
						Message: expectedFreshness.Validate().Error(),
					})

					Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
				})
			})
		})

		Context("when the request JSON is not valid", func() {
			var garbageRequest []byte

			BeforeEach(func() {
				garbageRequest = []byte(`garbage`)
				handler.Create(responseRecorder, newTestRequest(garbageRequest))
			})

			It("responds with 400 BAD REQUEST", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusBadRequest))
			})

			It("responds with a relevant error message", func() {
				err := json.Unmarshal(garbageRequest, &receptor.FreshDomainCreateRequest{})
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.InvalidJSON,
					Message: err.Error(),
				})

				Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
			})
		})
	})
})
