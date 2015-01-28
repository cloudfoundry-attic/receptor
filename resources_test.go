package receptor_test

import (
	"encoding/json"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"

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
			Context("with security group rules", func() {
				var expectedRequest receptor.TaskCreateRequest
				var payload string
				BeforeEach(func() {
					payload = `{
						"egress_rules":[
		          {
				        "protocol": "tcp",
								"destinations": ["0.0.0.0/0"],
				        "port_range": {
					        "start": 1,
					        "end": 1024
				        }
			        }
		        ]
					}`
					expectedRequest = receptor.TaskCreateRequest{
						EgressRules: []models.SecurityGroupRule{
							{
								Protocol:     "tcp",
								Destinations: []string{"0.0.0.0/0"},
								PortRange: &models.PortRange{
									Start: 1,
									End:   1024,
								},
							},
						},
					}
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
			Context("with security group rules", func() {
				var expectedResponse receptor.TaskResponse
				var payload string
				BeforeEach(func() {
					payload = `{
						"egress_rules":[
		          {
				        "protocol": "tcp",
								"destinations": ["0.0.0.0/0"],
				        "port_range": {
					        "start": 1,
					        "end": 1024
				        }
			        }
		        ]
					}`
					expectedResponse = receptor.TaskResponse{
						EgressRules: []models.SecurityGroupRule{
							{
								Protocol:     "tcp",
								Destinations: []string{"0.0.0.0/0"},
								PortRange: &models.PortRange{
									Start: 1,
									End:   1024,
								},
							},
						},
					}
				})

				It("unmarshals", func() {
					var actualRequest receptor.TaskResponse
					err := json.Unmarshal([]byte(payload), &actualRequest)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(actualRequest).Should(Equal(expectedResponse))
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
			Context("with security group rules", func() {
				var expectedRequest receptor.DesiredLRPCreateRequest
				var payload string

				BeforeEach(func() {
					payload = `{
						"egress_rules":[
		          {
				        "protocol": "tcp",
								"destinations": ["0.0.0.0/0"],
				        "ports": [80, 443],
				        "log": true
			        }
		        ]
					}`
					expectedRequest = receptor.DesiredLRPCreateRequest{
						EgressRules: []models.SecurityGroupRule{
							{
								Protocol:     "tcp",
								Destinations: []string{"0.0.0.0/0"},
								Ports:        []uint16{80, 443},
								Log:          true,
							},
						},
					}
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
			Context("with security group rules", func() {
				var expectedResponse receptor.DesiredLRPResponse
				var payload string

				BeforeEach(func() {
					payload = `{
						"egress_rules":[
		          {
				        "protocol": "tcp",
								"destinations": ["0.0.0.0/0"],
				        "port_range": {
					        "start": 1,
					        "end": 1024
				        }
			        }
		        ]
					}`
					expectedResponse = receptor.DesiredLRPResponse{
						EgressRules: []models.SecurityGroupRule{
							{
								Protocol:     "tcp",
								Destinations: []string{"0.0.0.0/0"},
								PortRange: &models.PortRange{
									Start: 1,
									End:   1024,
								},
							},
						},
					}
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

	Describe("RoutingInfo", func() {
		var r receptor.RoutingInfo

		BeforeEach(func() {
			r.Other = make(map[string]*json.RawMessage)
		})

		Context("Serialization", func() {
			jsonRoutes := `{
					"cf-router": [{ "port": 1, "hostnames": ["a", "b"]}],
					"foo" : "bar"
					}`

			Context("MarshalJson", func() {
				It("marshals routes when present", func() {
					r.CFRoutes = append(r.CFRoutes, receptor.CFRoute{Port: 1, Hostnames: []string{"a", "b"}})
					msg := json.RawMessage([]byte(`"bar"`))
					r.Other["foo"] = &msg
					bytes, err := json.Marshal(r)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(bytes).Should(MatchJSON(jsonRoutes))
				})
			})

			Context("Unmarshal", func() {
				It("returns both cf-routes and other", func() {
					err := json.Unmarshal([]byte(jsonRoutes), &r)
					Ω(err).ShouldNot(HaveOccurred())

					Ω(r.CFRoutes).Should(HaveLen(1))
					route := r.CFRoutes[0]
					Ω(route.Port).Should(Equal(uint16(1)))
					Ω(route.Hostnames).Should(ConsistOf("a", "b"))

					Ω(r.Other).Should(HaveLen(1))
					raw := r.Other["foo"]
					Ω([]byte(*raw)).Should(Equal([]byte(`"bar"`)))
				})
			})
		})
	})
})
