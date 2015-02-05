package receptor_test

import (
	"encoding/json"

	"github.com/cloudfoundry-incubator/receptor"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RoutingInfoHelpers", func() {
	var (
		route1 receptor.CFRoute
		route2 receptor.CFRoute
		route3 receptor.CFRoute

		routes receptor.CFRoutes
	)

	BeforeEach(func() {
		route1 = receptor.CFRoute{
			Hostnames: []string{"foo1.example.com", "bar1.examaple.com"},
			Port:      11111,
		}
		route2 = receptor.CFRoute{
			Hostnames: []string{"foo2.example.com", "bar2.examaple.com"},
			Port:      22222,
		}
		route3 = receptor.CFRoute{
			Hostnames: []string{"foo3.example.com", "bar3.examaple.com"},
			Port:      33333,
		}

		routes = receptor.CFRoutes{route1, route2, route3}
	})

	Describe("RoutingInfo", func() {
		var routingInfo receptor.RoutingInfo

		JustBeforeEach(func() {
			routingInfo = routes.RoutingInfo()
		})

		It("wraps the serialized routes with the correct key", func() {
			expectedBytes, err := json.Marshal(routes)
			Ω(err).ShouldNot(HaveOccurred())

			payload, err := routingInfo[receptor.CF_ROUTER].MarshalJSON()
			Ω(err).ShouldNot(HaveOccurred())

			Ω(payload).Should(MatchJSON(expectedBytes))
		})

		Context("when CFRoutes is empty", func() {
			BeforeEach(func() {
				routes = receptor.CFRoutes{}
			})

			It("marshals an empty list", func() {
				payload, err := routingInfo[receptor.CF_ROUTER].MarshalJSON()
				Ω(err).ShouldNot(HaveOccurred())

				Ω(payload).Should(MatchJSON(`[]`))
			})
		})
	})

	Describe("CFRoutesFromRoutingInfo", func() {
		var (
			routesResult    receptor.CFRoutes
			conversionError error

			routingInfo receptor.RoutingInfo
		)

		JustBeforeEach(func() {
			routesResult, conversionError = receptor.CFRoutesFromRoutingInfo(routingInfo)
		})

		Context("when CF routes are present in the routing info", func() {
			BeforeEach(func() {
				routingInfo = routes.RoutingInfo()
			})

			It("returns the routes", func() {
				Ω(routes).Should(Equal(routesResult))
			})

			Context("when the CF routes are nil", func() {
				BeforeEach(func() {
					routingInfo = receptor.RoutingInfo{receptor.CF_ROUTER: nil}
				})

				It("returns nil routes", func() {
					Ω(conversionError).ShouldNot(HaveOccurred())
					Ω(routesResult).Should(BeNil())
				})
			})
		})

		Context("when CF routes are not present in the routing info", func() {
			BeforeEach(func() {
				routingInfo = receptor.RoutingInfo{}
			})

			It("returns nil routes", func() {
				Ω(conversionError).ShouldNot(HaveOccurred())
				Ω(routesResult).Should(BeNil())
			})
		})

		Context("when the routing info is nil", func() {
			BeforeEach(func() {
				routingInfo = nil
			})

			It("returns nil routes", func() {
				Ω(conversionError).ShouldNot(HaveOccurred())
				Ω(routesResult).Should(BeNil())
			})
		})
	})
})
