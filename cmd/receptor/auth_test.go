package main_test

import (
	"net/http"

	"github.com/cloudfoundry-incubator/receptor/cmd/receptor/testrunner"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Basic Auth", func() {
	JustBeforeEach(func() {
		receptorProcess = ginkgomon.Invoke(receptorRunner)
	})

	AfterEach(func() {
		ginkgomon.Kill(receptorProcess)
	})

	Context("when a request without auth is made", func() {
		var res *http.Response
		JustBeforeEach(func() {
			var err error
			httpClient := new(http.Client)
			res, err = httpClient.Get("http://" + receptorAddress)
			Ω(err).ShouldNot(HaveOccurred())
			res.Body.Close()
		})

		Context("when the username and password have been set", func() {
			It("returns 401 for all requests", func() {
				Ω(res.StatusCode).Should(Equal(http.StatusUnauthorized))
			})
		})

		Context("and the username and password have not been set", func() {
			BeforeEach(func() {
				receptorArgs.Username = ""
				receptorArgs.Password = ""
				receptorRunner = testrunner.New(receptorBinPath, receptorArgs)
			})

			It("does not return 401", func() {
				Ω(res.StatusCode).Should(Equal(http.StatusNotFound))
			})
		})
	})
})
