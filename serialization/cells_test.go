package serialization_test

import (
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	"github.com/cloudfoundry-incubator/runtime-schema/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ExecutorPresence Serialization", func() {
	Describe("ExecutorPresenceToCellResponse", func() {
		var executorPresence models.ExecutorPresence

		BeforeEach(func() {
			executorPresence = models.ExecutorPresence{
				ExecutorID: "executor-id-0",
				Stack:      "stack-0",
			}
		})

		It("serializes all the fields", func() {
			expectedResponse := receptor.CellResponse{
				CellID: "executor-id-0",
				Stack:  "stack-0",
			}

			actualResponse := serialization.ExecutorPresenceToCellResponse(executorPresence)
			Ω(actualResponse).Should(Equal(expectedResponse))
		})
	})
})
