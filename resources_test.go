package receptor_test

import (
	"encoding/json"

	"github.com/cloudfoundry-incubator/receptor"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Resources", func() {
	Describe("TaskCreateRequest", func() {
		Describe("Unmarshalling", func() {
			Context("with invalid actions", func() {
				var expectedRequest receptor.TaskCreateRequest
				var payload string

				BeforeEach(func() {
					expectedRequest = receptor.TaskCreateRequest{}
				})

				Context("with null action", func() {
					BeforeEach(func() {
						payload = `{
              "action": null
            }`
					})

					It("unmarshals", func() {
						var actualRequest receptor.TaskCreateRequest
						err := json.Unmarshal([]byte(payload), &actualRequest)
						Ω(err).ShouldNot(HaveOccurred())
						Ω(actualRequest).Should(Equal(expectedRequest))
					})
				})

				Context("with missing action", func() {
					BeforeEach(func() {
						payload = `{}`
					})

					It("unmarshals", func() {
						var actualRequest receptor.TaskCreateRequest
						err := json.Unmarshal([]byte(payload), &actualRequest)
						Ω(err).ShouldNot(HaveOccurred())
						Ω(actualRequest).Should(Equal(expectedRequest))
					})
				})
			})
		})
	})

	Describe("TaskResponse", func() {
		Describe("Unmarshalling", func() {
			Context("with invalid actions", func() {
				var expectedResponse receptor.TaskResponse
				var payload string

				BeforeEach(func() {
					expectedResponse = receptor.TaskResponse{}
				})

				Context("with null action", func() {
					BeforeEach(func() {
						payload = `{
              "action": null
            }`
					})

					It("unmarshals", func() {
						var actualResponse receptor.TaskResponse
						err := json.Unmarshal([]byte(payload), &actualResponse)
						Ω(err).ShouldNot(HaveOccurred())
						Ω(actualResponse).Should(Equal(expectedResponse))
					})
				})

				Context("with missing action", func() {
					BeforeEach(func() {
						payload = `{}`
					})

					It("unmarshals", func() {
						var actualResponse receptor.TaskResponse
						err := json.Unmarshal([]byte(payload), &actualResponse)
						Ω(err).ShouldNot(HaveOccurred())
						Ω(actualResponse).Should(Equal(expectedResponse))
					})
				})
			})
		})
	})

	Describe("DesiredLRPCreateRequest", func() {
		Describe("Unmarshalling", func() {
			Context("with invalid actions", func() {
				var expectedRequest receptor.DesiredLRPCreateRequest
				var payload string

				BeforeEach(func() {
					expectedRequest = receptor.DesiredLRPCreateRequest{}
				})

				Context("with null action", func() {
					BeforeEach(func() {
						payload = `{
              "setup": null,
              "action": null,
              "monitor": null
            }`
					})

					It("unmarshals", func() {
						var actualRequest receptor.DesiredLRPCreateRequest
						err := json.Unmarshal([]byte(payload), &actualRequest)
						Ω(err).ShouldNot(HaveOccurred())
						Ω(actualRequest).Should(Equal(expectedRequest))
					})
				})

				Context("with missing action", func() {
					BeforeEach(func() {
						payload = `{}`
					})

					It("unmarshals", func() {
						var actualRequest receptor.DesiredLRPCreateRequest
						err := json.Unmarshal([]byte(payload), &actualRequest)
						Ω(err).ShouldNot(HaveOccurred())
						Ω(actualRequest).Should(Equal(expectedRequest))
					})
				})
			})
		})
	})

	Describe("DesiredLRPResponse", func() {
		Describe("Unmarshalling", func() {
			Context("with invalid actions", func() {
				var expectedResponse receptor.DesiredLRPResponse
				var payload string

				BeforeEach(func() {
					expectedResponse = receptor.DesiredLRPResponse{}
				})

				Context("with null action", func() {
					BeforeEach(func() {
						payload = `{
              "setup": null,
              "action": null,
              "monitor": null
            }`
					})

					It("unmarshals", func() {
						var actualResponse receptor.DesiredLRPResponse
						err := json.Unmarshal([]byte(payload), &actualResponse)
						Ω(err).ShouldNot(HaveOccurred())
						Ω(actualResponse).Should(Equal(expectedResponse))
					})
				})

				Context("with missing action", func() {
					BeforeEach(func() {
						payload = `{}`
					})

					It("unmarshals", func() {
						var actualResponse receptor.DesiredLRPResponse
						err := json.Unmarshal([]byte(payload), &actualResponse)
						Ω(err).ShouldNot(HaveOccurred())
						Ω(actualResponse).Should(Equal(expectedResponse))
					})
				})
			})
		})
	})
})
