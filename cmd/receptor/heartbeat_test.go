package main_test

import (
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Heartbeating", func() {
	BeforeEach(func() {
		receptorProcess = ginkgomon.Invoke(receptorRunner)
	})

	AfterEach(func() {
		ginkgomon.Kill(receptorProcess)
	})

	It("heartbeats its presence to the BBS with the task handler URL", func() {
		presence, err := bbs.Receptor()
		Ω(err).ShouldNot(HaveOccurred())

		Ω(presence.ReceptorID).ShouldNot(BeEmpty())
		Ω(presence.ReceptorURL).Should(Equal("http://" + receptorTaskHandlerAddress))
	})
})
