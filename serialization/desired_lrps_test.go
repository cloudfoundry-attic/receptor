package serialization_test

import (
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	"github.com/cloudfoundry-incubator/runtime-schema/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DesiredLRP Serialization", func() {
	Describe("DesiredLRPFromRequest", func() {
		var request receptor.DesiredLRPCreateRequest
		var desiredLRP models.DesiredLRP
		var securityRule models.SecurityGroupRule

		BeforeEach(func() {
			securityRule = models.SecurityGroupRule{
				Protocol: "tcp",
				PortRange: models.PortRange{
					Start: 1,
					End:   1024,
				},
				Destination: models.CIDR{
					NetworkAddress: "0.0.0.0",
					PrefixLength:   0,
				},
			}
			request = receptor.DesiredLRPCreateRequest{
				ProcessGuid: "the-process-guid",
				Domain:      "the-domain",
				Stack:       "the-stack",
				RootFSPath:  "the-rootfs-path",
				Annotation:  "foo",
				Instances:   1,
				Ports:       []uint32{2345, 6789},
				Action: &models.RunAction{
					Path: "the-path",
				},
				StartTimeout: 4,
				Privileged:   true,
				SecurityGroupRules: []models.SecurityGroupRule{
					securityRule,
				},
			}
		})
		JustBeforeEach(func() {
			desiredLRP = serialization.DesiredLRPFromRequest(request)
		})

		It("translates the request into a DesiredLRP model, preserving attributes", func() {
			Ω(desiredLRP.ProcessGuid).Should(Equal("the-process-guid"))
			Ω(desiredLRP.Domain).Should(Equal("the-domain"))
			Ω(desiredLRP.Stack).Should(Equal("the-stack"))
			Ω(desiredLRP.RootFSPath).Should(Equal("the-rootfs-path"))
			Ω(desiredLRP.Annotation).Should(Equal("foo"))
			Ω(desiredLRP.StartTimeout).Should(Equal(uint(4)))
			Ω(desiredLRP.Ports).Should(HaveLen(2))
			Ω(desiredLRP.Ports[0]).Should(Equal(uint32(2345)))
			Ω(desiredLRP.Ports[1]).Should(Equal(uint32(6789)))
			Ω(desiredLRP.Privileged).Should(BeTrue())
			Ω(desiredLRP.SecurityGroupRules).Should(HaveLen(1))
			Ω(desiredLRP.SecurityGroupRules[0].Protocol).Should(Equal(securityRule.Protocol))
			Ω(desiredLRP.SecurityGroupRules[0].PortRange).Should(Equal(securityRule.PortRange))
			Ω(desiredLRP.SecurityGroupRules[0].Destination).Should(Equal(securityRule.Destination))
		})
	})

	Describe("DesiredLRPToResponse", func() {
		var desiredLRP models.DesiredLRP
		var securityRule models.SecurityGroupRule

		BeforeEach(func() {
			securityRule = models.SecurityGroupRule{
				Protocol: "tcp",
				PortRange: models.PortRange{
					Start: 1,
					End:   1024,
				},
				Destination: models.CIDR{
					NetworkAddress: "0.0.0.0",
					PrefixLength:   0,
				},
			}

			desiredLRP = models.DesiredLRP{
				ProcessGuid: "process-guid-0",
				Domain:      "domain-0",
				RootFSPath:  "root-fs-path-0",
				Instances:   127,
				Stack:       "stack-0",
				EnvironmentVariables: []models.EnvironmentVariable{
					{Name: "ENV_VAR_NAME", Value: "value"},
				},
				Action:       &models.RunAction{Path: "/bin/true"},
				StartTimeout: 4,
				DiskMB:       126,
				MemoryMB:     1234,
				CPUWeight:    192,
				Privileged:   true,
				Ports: []uint32{
					456,
				},
				Routes:     []string{"route-0", "route-1"},
				LogGuid:    "log-guid-0",
				LogSource:  "log-source-name-0",
				Annotation: "annotation-0",
				SecurityGroupRules: []models.SecurityGroupRule{
					securityRule,
				},
			}
		})

		It("serializes all the fields", func() {
			expectedResponse := receptor.DesiredLRPResponse{
				ProcessGuid: "process-guid-0",
				Domain:      "domain-0",
				RootFSPath:  "root-fs-path-0",
				Instances:   127,
				Stack:       "stack-0",
				EnvironmentVariables: []receptor.EnvironmentVariable{
					{Name: "ENV_VAR_NAME", Value: "value"},
				},
				Action:       &models.RunAction{Path: "/bin/true"},
				StartTimeout: 4,
				DiskMB:       126,
				MemoryMB:     1234,
				CPUWeight:    192,
				Privileged:   true,
				Ports: []uint32{
					456,
				},
				Routes:     []string{"route-0", "route-1"},
				LogGuid:    "log-guid-0",
				LogSource:  "log-source-name-0",
				Annotation: "annotation-0",
				SecurityGroupRules: []models.SecurityGroupRule{
					securityRule,
				},
			}

			actualResponse := serialization.DesiredLRPToResponse(desiredLRP)
			Ω(actualResponse).Should(Equal(expectedResponse))
		})
	})
})
