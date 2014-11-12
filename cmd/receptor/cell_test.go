package main_test

import (
	"time"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cell API", func() {
	var heartbeatProcess ifrit.Process
	var executorPresence models.ExecutorPresence
	var heartbeatInterval time.Duration

	BeforeEach(func() {
		heartbeatInterval = 100 * time.Millisecond
		executorPresence = models.ExecutorPresence{
			ExecutorID: "cell-0",
			Stack:      "stack-0",
		}
		heartbeatRunner := bbs.NewExecutorHeartbeat(executorPresence, heartbeatInterval)
		heartbeatProcess = ginkgomon.Invoke(heartbeatRunner)
		receptorProcess = ginkgomon.Invoke(receptorRunner)
	})

	AfterEach(func() {
		ginkgomon.Kill(receptorProcess)
		ginkgomon.Kill(heartbeatProcess)
	})

	Describe("GET /cells", func() {
		var cellResponses []receptor.CellResponse
		var getErr error

		BeforeEach(func() {
			Eventually(func() []models.ExecutorPresence {
				executorPresences, err := bbs.GetAllExecutors()
				Ω(err).ShouldNot(HaveOccurred())
				return executorPresences
			}).Should(HaveLen(1))

			cellResponses, getErr = client.Cells()
		})

		It("responds without error", func() {
			Ω(getErr).ShouldNot(HaveOccurred())
		})

		It("has the correct data from the bbs", func() {
			executorPresences, err := bbs.GetAllExecutors()
			Ω(err).ShouldNot(HaveOccurred())

			expectedResponses := make([]receptor.CellResponse, 0, 1)
			for _, executorPresence := range executorPresences {
				expectedResponses = append(expectedResponses, serialization.ExecutorPresenceToCellResponse(executorPresence))
			}

			Ω(cellResponses).Should(ConsistOf(expectedResponses))
		})

	})

})
