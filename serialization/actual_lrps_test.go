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
					"address-0",
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
				Address:      "address-0",
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

	Describe("ActualLRPFromResponse", func() {
		var actualLRPResponse receptor.ActualLRPResponse

		BeforeEach(func() {
			actualLRPResponse = receptor.ActualLRPResponse{
				ProcessGuid:  "process-guid-0",
				InstanceGuid: "instance-guid",
				CellID:       "cell-id",
				Domain:       "domain",
				Index:        0,
				Address:      "address",
				Ports:        []receptor.PortMapping{{ContainerPort: 10000, HostPort: 10000}},
				State:        receptor.ActualLRPStateRunning,
				Since:        99999999999,
			}
		})

		It("deserializes all the fields", func() {
			actualLRP := serialization.ActualLRPFromResponse(actualLRPResponse)
			Ω(actualLRP).Should(Equal(models.ActualLRP{
				ActualLRPKey:          models.NewActualLRPKey("process-guid-0", 0, "domain"),
				ActualLRPContainerKey: models.NewActualLRPContainerKey("instance-guid", "cell-id"),
				ActualLRPNetInfo:      models.NewActualLRPNetInfo("address", []models.PortMapping{{ContainerPort: 10000, HostPort: 10000}}),
				State:                 models.ActualLRPStateRunning,
				Since:                 99999999999,
			}))
		})

		It("maps receptor states to model states", func() {
			expectedStateMap := map[receptor.ActualLRPState]models.ActualLRPState{
				receptor.ActualLRPStateUnclaimed: models.ActualLRPStateUnclaimed,
				receptor.ActualLRPStateClaimed:   models.ActualLRPStateClaimed,
				receptor.ActualLRPStateRunning:   models.ActualLRPStateRunning,
			}

			for jsonState, modelState := range expectedStateMap {
				actualLRPResponse.State = jsonState
				Ω(serialization.ActualLRPFromResponse(actualLRPResponse).State).Should(Equal(modelState))
			}

			actualLRPResponse.State = ""
			Ω(serialization.ActualLRPFromResponse(actualLRPResponse).State).Should(Equal(models.ActualLRPState("")))
		})
	})
})
