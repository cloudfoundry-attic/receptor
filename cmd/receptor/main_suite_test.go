package main_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

const username = "username"
const password = "password"

var receptorBinPath string
var receptorAddress string
var etcdPort int

var _ = SynchronizedBeforeSuite(
	func() []byte {
		receptorConfig, err := gexec.Build("github.com/cloudfoundry-incubator/receptor/cmd/receptor", "-race")
		Ω(err).ShouldNot(HaveOccurred())
		return []byte(receptorConfig)
	},
	func(receptorConfig []byte) {
		receptorBinPath = string(receptorConfig)
		receptorAddress = fmt.Sprintf("127.0.0.1:%d", 6700+GinkgoParallelNode())
		etcdPort = 4001 + GinkgoParallelNode()
	},
)

var _ = SynchronizedAfterSuite(func() {
}, func() {
	gexec.CleanupBuildArtifacts()
})

func TestReceptor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Receptor Suite")
}
