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

	Describe("Bump", func() {
		Context("with a structured request", func() {
			var freshDomainBumpRequest receptor.FreshDomainBumpRequest
			var expectedFreshness models.Freshness

			BeforeEach(func() {
				freshDomainBumpRequest = receptor.FreshDomainBumpRequest{
					Domain:       "domain-1",
					TTLInSeconds: 1000,
				}

				expectedFreshness = models.Freshness{
					Domain:       "domain-1",
					TTLInSeconds: 1000,
				}
			})

			JustBeforeEach(func() {
				handler.Bump(responseRecorder, newTestRequest(freshDomainBumpRequest))
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
				var validationError = models.ValidationError{}

				BeforeEach(func() {
					fakeBBS.BumpFreshnessReturns(validationError)
				})

				It("responds with 400 BAD REQUEST", func() {
					Ω(responseRecorder.Code).Should(Equal(http.StatusBadRequest))
				})

				It("responds with a relevant error message", func() {
					expectedBody, _ := json.Marshal(receptor.Error{
						Type:    receptor.InvalidFreshness,
						Message: validationError.Error(),
					})

					Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
				})
			})
		})

		Context("when the request JSON is not valid", func() {
			var garbageRequest []byte

			BeforeEach(func() {
				garbageRequest = []byte(`garbage`)
				handler.Bump(responseRecorder, newTestRequest(garbageRequest))
			})

			It("responds with 400 BAD REQUEST", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusBadRequest))
			})

			It("responds with a relevant error message", func() {
				err := json.Unmarshal(garbageRequest, &receptor.FreshDomainBumpRequest{})
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.InvalidJSON,
					Message: err.Error(),
				})

				Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
			})
		})
	})

	Describe("GetAll", func() {
		var freshnesses []models.Freshness

		BeforeEach(func() {
			freshnesses = []models.Freshness{
				{
					Domain:       "domain-a",
					TTLInSeconds: 10,
				},
				{
					Domain:       "domain-b",
					TTLInSeconds: 30,
				},
			}
		})

		JustBeforeEach(func() {
			handler.GetAll(responseRecorder, newTestRequest(""))
		})

		Context("when reading freshnesses from BBS succeeds", func() {
			BeforeEach(func() {
				fakeBBS.FreshnessesReturns(freshnesses, nil)
			})

			It("call the BBS to retrieve the actual LRPs", func() {
				Ω(fakeBBS.FreshnessesCallCount()).Should(Equal(1))
			})

			It("responds with 200 Status OK", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusOK))
			})

			It("returns a list of fresh domain responses", func() {
				response := []receptor.FreshDomainResponse{}
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(response).Should(HaveLen(2))
				for _, freshness := range freshnesses {
					Ω(response).Should(ContainElement(serialization.FreshnessToResponse(freshness)))
				}
			})
		})

		Context("when the BBS returns no freshnesses", func() {
			BeforeEach(func() {
				fakeBBS.FreshnessesReturns([]models.Freshness{}, nil)
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
				fakeBBS.FreshnessesReturns([]models.Freshness{}, errors.New("Something went wrong"))
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
