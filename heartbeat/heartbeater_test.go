package heartbeat_test

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/apcera/nats"
	"github.com/cloudfoundry-incubator/receptor/heartbeat"
	"github.com/cloudfoundry/gunk/diegonats"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Heartbeater", func() {
	var fakeNatsClient *diegonats.FakeNATSClient
	var heartbeater *heartbeat.Heartbeater
	var heartbeaterProcess ifrit.Process
	var registrations chan heartbeat.RegistryMessage
	var initialRegisterInterval = 100 * time.Millisecond
	var expectedRegistryMsg = heartbeat.RegistryMessage{
		URIs: []string{"foo.bar.com", "example.com"},
		Host: "1.2.3.4",
		Port: 4567,
	}

	BeforeEach(func() {
		logger := lager.NewLogger("test")
		fakeNatsClient = diegonats.NewFakeClient()
		heartbeater = heartbeat.New(fakeNatsClient, expectedRegistryMsg, initialRegisterInterval, logger)
		registrations = make(chan heartbeat.RegistryMessage, 1)

		fakeNatsClient.Subscribe("router.register", func(msg *nats.Msg) {
			registration := heartbeat.RegistryMessage{}
			fromJson(msg.Data, &registration)
			registrations <- registration
		})
	})

	JustBeforeEach(func() {
		heartbeaterProcess = ifrit.Invoke(heartbeater)
	})

	AfterEach(func() {
		ginkgomon.Kill(heartbeaterProcess)
	})

	Context("when the router greeting is successful", func() {
		var expectedRegisterInterval = time.Second
		var greetingMsg = heartbeat.GreetingMessage{
			Id:               "some-router-id",
			Hosts:            []string{"1.2.3.4"},
			RegisterInterval: expectedRegisterInterval,
		}

		BeforeEach(func() {
			fakeNatsClient.Subscribe("router.greet", func(msg *nats.Msg) {
				fakeNatsClient.Publish(msg.Reply, toJson(greetingMsg))
			})
		})

		Context("when no new router.start messages occur", func() {
			It("emits a registration on the the 'router.register' topic at the interval specified in the greeting", func() {
				testHeartbeatsOnTime(registrations, expectedRegistryMsg, expectedRegisterInterval)
			})
		})

		Context("when a router.start message occurs with a different register interval", func() {
			var newExpectedRegisterInterval = expectedRegisterInterval * 2
			var newGreetingMsg = heartbeat.GreetingMessage{
				Id:               "some-router-id",
				Hosts:            []string{"1.2.3.4"},
				RegisterInterval: newExpectedRegisterInterval,
			}

			JustBeforeEach(func() {
				fakeNatsClient.Publish("router.start", toJson(newGreetingMsg))
			})

			It("updates the heartbeat interval", func() {
				testHeartbeatsOnTime(registrations, expectedRegistryMsg, newExpectedRegisterInterval)
			})
		})
	})

	Context("when the router greeting is not successful", func() {
		It("emits a registration on the the 'router.register' topic at the initial interval specified", func() {
			testHeartbeatsOnTime(registrations, expectedRegistryMsg, initialRegisterInterval)
		})
	})
})

const numHeartbeats = 3
const marginOfError = 50 * time.Millisecond

func testHeartbeatsOnTime(
	registrations chan heartbeat.RegistryMessage,
	expectedRegistryMsg heartbeat.RegistryMessage,
	expectedInterval time.Duration,
) {
	msg := heartbeat.RegistryMessage{}

	Eventually(registrations, marginOfError).Should(Receive(&msg))

	for i := 1; i <= numHeartbeats-1; i++ {
		By(fmt.Sprint("Waiting for registration ", i, " of ", numHeartbeats))

		Consistently(registrations, expectedInterval-marginOfError).ShouldNot(Receive())
		Eventually(registrations, marginOfError*2).Should(Receive(&msg))
		Ω(msg).Should(Equal(expectedRegistryMsg))

		By(fmt.Sprint("Registration ", i, " of ", numHeartbeats, " received"))
	}
}

func toJson(obj interface{}) []byte {
	jsonBytes, err := json.Marshal(obj)
	Ω(err).ShouldNot(HaveOccurred())
	return jsonBytes
}

func fromJson(jsonBytes []byte, obj interface{}) {
	err := json.Unmarshal(jsonBytes, obj)
	Ω(err).ShouldNot(HaveOccurred())
}
