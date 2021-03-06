package main_test

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/cloudfoundry-incubator/bbs"
	"github.com/cloudfoundry-incubator/consuladapter"
	"github.com/cloudfoundry-incubator/consuladapter/consulrunner"
	"github.com/cloudfoundry-incubator/receptor"

	bbstestrunner "github.com/cloudfoundry-incubator/bbs/cmd/bbs/testrunner"
	"github.com/cloudfoundry-incubator/receptor/cmd/receptor/testrunner"
	"github.com/cloudfoundry/gunk/diegonats"
	"github.com/cloudfoundry/gunk/diegonats/gnatsdrunner"
	"github.com/cloudfoundry/storeadapter"
	"github.com/cloudfoundry/storeadapter/storerunner/etcdstorerunner"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"
	"github.com/tedsuo/ifrit/grouper"

	"testing"
	"time"
)

const (
	username = "username"
	password = "password"
)

var natsPort int
var natsAddress string
var natsClient diegonats.NATSClient
var natsServerRunner *ginkgomon.Runner
var natsClientRunner diegonats.NATSClientRunner
var natsGroupProcess ifrit.Process

var etcdPort int
var etcdUrl string
var etcdRunner *etcdstorerunner.ETCDClusterRunner
var etcdAdapter storeadapter.StoreAdapter

var consulRunner *consulrunner.ClusterRunner
var consulSession *consuladapter.Session

var bbsArgs bbstestrunner.Args
var bbsBinPath string
var bbsURL *url.URL
var bbsRunner *ginkgomon.Runner
var bbsProcess ifrit.Process
var bbsClient bbs.Client

var serviceClient bbs.ServiceClient

var logger lager.Logger

var client receptor.Client
var receptorBinPath string
var receptorAddress string
var receptorArgs testrunner.Args
var receptorRunner *ginkgomon.Runner
var receptorProcess ifrit.Process

func TestReceptor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Receptor Cmd Suite")
}

var _ = SynchronizedBeforeSuite(
	func() []byte {
		bbsConfig, err := gexec.Build("github.com/cloudfoundry-incubator/bbs/cmd/bbs", "-race")
		Expect(err).NotTo(HaveOccurred())

		receptorConfig, err := gexec.Build("github.com/cloudfoundry-incubator/receptor/cmd/receptor", "-race")
		Expect(err).NotTo(HaveOccurred())

		return []byte(strings.Join([]string{receptorConfig, bbsConfig}, ","))
	},
	func(pathsByte []byte) {
		path := string(pathsByte)
		receptorBinPath = strings.Split(path, ",")[0]
		bbsBinPath = strings.Split(path, ",")[1]
		SetDefaultEventuallyTimeout(15 * time.Second)

		etcdPort = 4001 + GinkgoParallelNode()
		etcdUrl = fmt.Sprintf("http://127.0.0.1:%d", etcdPort)
		etcdRunner = etcdstorerunner.NewETCDClusterRunner(etcdPort, 1, nil)

		consulRunner = consulrunner.NewClusterRunner(
			9001+config.GinkgoConfig.ParallelNode*consulrunner.PortOffsetLength,
			1,
			"http",
		)

		etcdRunner.Start()
		consulRunner.Start()
		consulRunner.WaitUntilReady()

		bbsAddress := fmt.Sprintf("127.0.0.1:%d", 13000+GinkgoParallelNode())

		bbsURL = &url.URL{
			Scheme: "http",
			Host:   bbsAddress,
		}

		bbsClient = bbs.NewClient(bbsURL.String())

		bbsArgs = bbstestrunner.Args{
			Address:           bbsAddress,
			AdvertiseURL:      bbsURL.String(),
			AuctioneerAddress: "some-address",
			EtcdCluster:       etcdUrl,
			ConsulCluster:     consulRunner.ConsulCluster(),

			EncryptionKeys: []string{"label:key"},
			ActiveKeyLabel: "label",
		}
	})

var _ = SynchronizedAfterSuite(func() {
	etcdRunner.Stop()
	consulRunner.Stop()
}, func() {
	gexec.CleanupBuildArtifacts()
})

var _ = BeforeEach(func() {
	logger = lagertest.NewTestLogger("test")

	etcdRunner.Reset()

	consulRunner.Reset()
	consulSession = consulRunner.NewSession("a-session")

	receptorAddress = fmt.Sprintf("127.0.0.1:%d", 6700+GinkgoParallelNode())

	etcdAdapter = etcdRunner.Adapter(nil)
	serviceClient = bbs.NewServiceClient(consulSession, clock.NewClock())

	natsPort = 4051 + GinkgoParallelNode()
	natsAddress = fmt.Sprintf("127.0.0.1:%d", natsPort)
	natsClient = diegonats.NewClient()
	natsGroupProcess = ginkgomon.Invoke(newNatsGroup())

	bbsRunner = bbstestrunner.New(bbsBinPath, bbsArgs)
	bbsProcess = ginkgomon.Invoke(bbsRunner)

	receptorURL := &url.URL{
		Scheme: "http",
		Host:   receptorAddress,
		User:   url.UserPassword(username, password),
	}

	client = receptor.NewClient(receptorURL.String())

	receptorArgs = testrunner.Args{
		RegisterWithRouter: true,
		DomainNames:        "example.com",
		Address:            receptorAddress,
		Username:           username,
		Password:           password,
		NatsAddresses:      natsAddress,
		NatsUsername:       "nats",
		NatsPassword:       "nats",
		ConsulCluster:      consulRunner.ConsulCluster(),
		BBSAddress:         bbsURL.String(),
	}
	receptorRunner = testrunner.New(receptorBinPath, receptorArgs)
})

var _ = AfterEach(func() {
	etcdAdapter.Disconnect()
	ginkgomon.Kill(natsGroupProcess)
	ginkgomon.Kill(bbsProcess)
})

func newNatsGroup() ifrit.Runner {
	natsServerRunner = gnatsdrunner.NewGnatsdTestRunner(natsPort)
	natsClientRunner = diegonats.NewClientRunner(natsAddress, "", "", logger, natsClient)
	return grouper.NewOrdered(os.Kill, grouper.Members{
		{"natsServer", natsServerRunner},
		{"natsClient", natsClientRunner},
	})
}
