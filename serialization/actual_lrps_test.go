package serialization_test

import (
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	"github.com/cloudfoundry-incubator/runtime-schema/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ActualLRP Serialization", func() {
	Describe("ActualLRPToResponse", func() {
		var actualLRP models.ActualLRP
		BeforeEach(func() {
			actualLRP = models.ActualLRP{
				ActualLRPKey: models.NewActualLRPKey(
					"process-guid-0",
					3,
					"some-domain",
				),
				ActualLRPContainerKey: models.NewActualLRPContainerKey(
					"instance-guid-0",
					"cell-id-0",
				),
				ActualLRPNetInfo: models.NewActualLRPNetInfo(
					"host-0",
					[]models.PortMapping{
						{
							ContainerPort: 2345,
							HostPort:      9876,
						},
					},
				),
				State: models.ActualLRPStateRunning,
				Since: 99999999999,
			}
		})

		It("serializes all the fields", func() {
			expectedResponse := receptor.ActualLRPResponse{
				ProcessGuid:  "process-guid-0",
				InstanceGuid: "instance-guid-0",
				CellID:       "cell-id-0",
				Domain:       "some-domain",
				Index:        3,
				Host:         "host-0",
				Ports: []receptor.PortMapping{
					{
						ContainerPort: 2345,
						HostPort:      9876,
					},
				},
				State: receptor.ActualLRPStateRunning,
				Since: 99999999999,
			}

			actualResponse := serialization.ActualLRPToResponse(actualLRP)
			Ω(actualResponse).Should(Equal(expectedResponse))
		})

		It("maps model states to receptor states", func() {
			expectedStateMap := map[models.ActualLRPState]receptor.ActualLRPState{
				models.ActualLRPStateUnclaimed: receptor.ActualLRPStateUnclaimed,
				models.ActualLRPStateClaimed:   receptor.ActualLRPStateClaimed,
				models.ActualLRPStateRunning:   receptor.ActualLRPStateRunning,
			}

			for modelState, jsonState := range expectedStateMap {
				actualLRP.State = modelState
				Ω(serialization.ActualLRPToResponse(actualLRP).State).Should(Equal(jsonState))
			}

			actualLRP.State = ""
			Ω(serialization.ActualLRPToResponse(actualLRP).State).Should(Equal(receptor.ActualLRPStateInvalid))
		})
	})
})
