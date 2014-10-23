package main_test

import (
	"github.com/cloudfoundry-incubator/receptor/cmd/receptor/testrunner"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Flag Validation", func() {
	JustBeforeEach(func() {
		receptorProcess = ifrit.Background(receptorRunner)
	})

	AfterEach(func() {
		ginkgomon.Kill(receptorProcess)
	})

	Context("when the nats address is not specified but the nats username is", func() {
		BeforeEach(func() {
			receptorArgs.NatsAddresses = ""
			receptorArgs.NatsUsername = "nats"
			receptorArgs.NatsPassword = ""
			receptorRunner = testrunner.New(receptorBinPath, receptorArgs)
		})

		It("exits with a non zero error code if nats user name is provided and nats address is not", func() {
			Eventually(receptorRunner).Should(gexec.Exit(1))
		})
	})

	Context("when the nats address is not specified but the nats password is", func() {
		BeforeEach(func() {
			receptorArgs.NatsAddresses = ""
			receptorArgs.NatsUsername = ""
			receptorArgs.NatsPassword = "nats"
			receptorRunner = testrunner.New(receptorBinPath, receptorArgs)
		})

		It("exits with a non zero error code if nats user name is provided and nats address is not", func() {
			Eventually(receptorRunner).Should(gexec.Exit(1))
		})
	})

	Context("when the nats info is specified but the domains are not", func() {
		BeforeEach(func() {
			receptorArgs.DomainNames = ""
			receptorRunner = testrunner.New(receptorBinPath, receptorArgs)
		})

		It("exits with a non zero error code if nats user name is provided and nats address is not", func() {
			Eventually(receptorRunner).Should(gexec.Exit(1))
		})
	})
})
