package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/cf-debug-server"
	"github.com/cloudfoundry-incubator/cf-lager"
	"github.com/cloudfoundry-incubator/receptor/handlers"
	"github.com/cloudfoundry-incubator/receptor/heartbeat"
	"github.com/cloudfoundry-incubator/receptor/task_watcher"
	Bbs "github.com/cloudfoundry-incubator/runtime-schema/bbs"
	"github.com/cloudfoundry/gunk/timeprovider"
	"github.com/cloudfoundry/gunk/workpool"
	"github.com/cloudfoundry/storeadapter/etcdstoreadapter"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/http_server"
	"github.com/tedsuo/ifrit/sigmon"
)

var serverDomainNames = flag.String(
	"domainNames",
	"",
	"Comma separated list of domains that should route to this server",
)

var serverAddress = flag.String(
	"address",
	"",
	"the host:port that the server is bound to",
)

var etcdCluster = flag.String(
	"etcdCluster",
	"http://127.0.0.1:4001",
	"comma-separated list of etcd addresses (http://ip:port)",
)

var username = flag.String(
	"username",
	"",
	"username for basic auth, enables basic auth if set",
)

var password = flag.String(
	"password",
	"",
	"password for basic auth",
)

var natsAddresses = flag.String(
	"natsAddresses",
	"",
	"comma-separated list of NATS addresses (ip:port)",
)

var natsUsername = flag.String(
	"natsUsername",
	"",
	"Username to connect to nats",
)

var natsPassword = flag.String(
	"natsPassword",
	"",
	"Password for nats user",
)

var initialHeartbeatInterval = flag.Duration(
	"initialHeartbeatInterval",
	time.Second,
	"Heartbeat interval to use prior to router greeting",
)

func main() {
	flag.Parse()

	cf_debug_server.Run()

	logger := cf_lager.New("receptor")
	logger.Info("starting")

	if err := validateNatsArguments(); err != nil {
		logger.Error("invalid-nats-flags", err)
		os.Exit(1)
	}

	bbs := initializeReceptorBBS(logger)

	handler := handlers.New(bbs, logger, *username, *password)

	members := grouper.Members{
		{"server", http_server.New(*serverAddress, handler)},
		{"watcher", task_watcher.New(bbs, logger)},
	}

	if shouldRegisterRoutes() {
		registration := initializeServerRegistration(logger)

		members = append(members, grouper.Member{
			Name:   "background_heartbeat",
			Runner: heartbeat.NewBackgroundHeartbeat(*natsAddresses, *natsUsername, *natsPassword, logger, registration),
		})
	}

	group := grouper.NewOrdered(os.Interrupt, members)

	monitor := ifrit.Invoke(sigmon.New(group))

	logger.Info("started")

	err := <-monitor.Wait()
	if err != nil {
		logger.Error("exited-with-failure", err)
		os.Exit(1)
	}

	logger.Info("exited")
	os.Exit(0) // FIXME: why am I needed?
}

func validateNatsArguments() error {
	if *natsAddresses == "" && (*natsUsername != "" || *natsPassword != "") {
		return errors.New("nats username/password set, but nats address was blank")
	}
	if *natsAddresses != "" && *serverDomainNames == "" {
		return errors.New("nats address specified, but no domain names to register were specified")
	}
	return nil
}

func shouldRegisterRoutes() bool {
	return *natsAddresses != ""
}

func initializeReceptorBBS(logger lager.Logger) Bbs.ReceptorBBS {
	etcdAdapter := etcdstoreadapter.NewETCDStoreAdapter(
		strings.Split(*etcdCluster, ","),
		workpool.NewWorkPool(10),
	)

	err := etcdAdapter.Connect()
	if err != nil {
		logger.Fatal("failed-to-connect-to-etcd", err)
	}

	return Bbs.NewReceptorBBS(etcdAdapter, timeprovider.NewTimeProvider(), logger)
}

func initializeServerRegistration(logger lager.Logger) (registration heartbeat.RegistryMessage) {
	domains := strings.Split(*serverDomainNames, ",")

	addressComponents := strings.Split(*serverAddress, ":")
	if len(addressComponents) != 2 {
		logger.Error("server-address-invalid", fmt.Errorf("%s is not a valid serverAddress", *serverAddress))
		os.Exit(1)
	}

	host := addressComponents[0]

	port, err := strconv.Atoi(addressComponents[1])
	if err != nil {
		logger.Error("server-address-invalid", fmt.Errorf("%s does not have a valid port", *serverAddress))
		os.Exit(1)
	}

	return heartbeat.RegistryMessage{
		URIs: domains,
		Host: host,
		Port: port,
	}
}
