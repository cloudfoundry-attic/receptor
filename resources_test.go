package receptor_test

import (
	"encoding/json"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/route-emitter/cfroutes"
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
		const jsonRoutes = `{
			"cf-router": [{ "port": 1, "hostnames": ["a", "b"]}],
			"foo" : "bar"
		}`

		var routingInfo receptor.RoutingInfo

		BeforeEach(func() {
			routingInfo = receptor.RoutingInfo{}
		})

		Describe("MarshalJson", func() {
			It("marshals routes when present", func() {
				routingInfo := receptor.RoutingInfo{}

				bytes, err := json.Marshal(cfroutes.CFRoutes{
					{Hostnames: []string{"a", "b"}, Port: 1},
				})
				Ω(err).ShouldNot(HaveOccurred())

				cfRawMessage := json.RawMessage(bytes)
				routingInfo[cfroutes.CF_ROUTER] = &cfRawMessage

				fooRawMessage := json.RawMessage([]byte(`"bar"`))
				routingInfo["foo"] = &fooRawMessage

				bytes, err = json.Marshal(routingInfo)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(bytes).Should(MatchJSON(jsonRoutes))
			})

			Context("when the routing info is empty", func() {
				It("serialization can omit it", func() {
					type wrapper struct {
						Name   string               `json:"name"`
						Routes receptor.RoutingInfo `json:"routes,omitempty"`
					}

					w := wrapper{
						Name:   "wrapper",
						Routes: receptor.RoutingInfo{},
					}

					Ω(w.Routes).ShouldNot(BeNil())
					Ω(w.Routes).Should(BeEmpty())

					bytes, err := json.Marshal(w)
					Ω(err).ShouldNot(HaveOccurred())

					Ω(bytes).Should(MatchJSON(`{"name":"wrapper"}`))
				})
			})
		})

		Describe("Unmarshal", func() {
			It("returns both cf-routes and other", func() {
				err := json.Unmarshal([]byte(jsonRoutes), &routingInfo)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(len(routingInfo)).Should(Equal(2))
				Ω(string(*routingInfo["cf-router"])).Should(MatchJSON(`[{"port": 1,"hostnames":["a", "b"]}]`))
				Ω(string(*routingInfo["foo"])).Should(MatchJSON(`"bar"`))
			})
		})
	})
})
