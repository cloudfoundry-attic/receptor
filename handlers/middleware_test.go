package handlers_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/handlers"
	"github.com/cloudfoundry-incubator/receptor/handlers/handler_fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Middleware", func() {
	var handler http.Handler
	var wrappedHandler *handler_fakes.FakeHandler
	var req *http.Request
	var res *httptest.ResponseRecorder

	BeforeEach(func() {
		req = newTestRequest("")
		res = httptest.NewRecorder()
		wrappedHandler = new(handler_fakes.FakeHandler)
	})

	Describe("BasicAuthWrap", func() {
		var expectedUsername = "user"
		var expectedPassword = "pass"

		BeforeEach(func() {
			handler = handlers.BasicAuthWrap(wrappedHandler, expectedUsername, expectedPassword)
		})

		Context("when the correct credentials are provided", func() {
			BeforeEach(func() {
				req.SetBasicAuth(expectedUsername, expectedPassword)
				handler.ServeHTTP(res, req)
			})

			It("calls the wrapped handler", func() {
				Ω(wrappedHandler.ServeHTTPCallCount()).Should(Equal(1))
			})
		})

		Context("when no credentials are provided", func() {
			BeforeEach(func() {
				handler.ServeHTTP(res, req)
			})

			It("doesn't call the wrapped handler", func() {
				Ω(wrappedHandler.ServeHTTPCallCount()).Should(Equal(0))
			})
		})

		Context("when incorrect credentials are provided", func() {
			BeforeEach(func() {
				req.SetBasicAuth(expectedUsername, "badPassword")
				handler.ServeHTTP(res, req)
			})

			It("returns 401 UNAUTHORIZED", func() {
				Ω(res.Code).Should(Equal(http.StatusUnauthorized))
			})

			It("returns an unauthorized error response", func() {
				expectedError := receptor.ErrorResponse{
					Error: http.StatusText(http.StatusUnauthorized),
				}
				expectedBody := expectedError.JSONReader()
				Ω(res.Body.String()).Should(Equal(expectedBody.String()))
			})

			It("doesn't call the wrapped handler", func() {
				Ω(wrappedHandler.ServeHTTPCallCount()).Should(Equal(0))
			})
		})
	})
})
