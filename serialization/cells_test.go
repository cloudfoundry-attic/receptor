package serialization_test

import (
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	"github.com/cloudfoundry-incubator/runtime-schema/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CellPresence Serialization", func() {
	Describe("CellPresenceToCellResponse", func() {
		var cellPresence models.CellPresence

		BeforeEach(func() {
			cellPresence = models.NewCellPresence("cell-id-0", "stack-0", "1.2.3.4", "the-zone")
		})

		It("serializes all the fields", func() {
			expectedResponse := receptor.CellResponse{
				CellID: "cell-id-0",
				Stack:  "stack-0",
			}

			actualResponse := serialization.CellPresenceToCellResponse(cellPresence)
			Ω(actualResponse).Should(Equal(expectedResponse))
		})
	})
})
