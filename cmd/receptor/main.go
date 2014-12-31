package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	cf_debug_server "github.com/cloudfoundry-incubator/cf-debug-server"
	cf_lager "github.com/cloudfoundry-incubator/cf-lager"
	"github.com/cloudfoundry-incubator/natbeat"
	"github.com/cloudfoundry-incubator/receptor/handlers"
	"github.com/cloudfoundry-incubator/receptor/task_handler"
	Bbs "github.com/cloudfoundry-incubator/runtime-schema/bbs"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/cloudfoundry/dropsonde"
	"github.com/cloudfoundry/gunk/diegonats"
	"github.com/cloudfoundry/gunk/timeprovider"
	"github.com/cloudfoundry/gunk/workpool"
	"github.com/cloudfoundry/storeadapter/etcdstoreadapter"
	"github.com/nu7hatch/gouuid"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/localip"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/http_server"
	"github.com/tedsuo/ifrit/sigmon"
)

var registerWithRouter = flag.Bool(
	"registerWithRouter",
	false,
	"Register this receptor instance with the router.",
)

var serverDomainNames = flag.String(
	"domainNames",
	"",
	"Comma separated list of domains that should route to this server.",
)

var serverAddress = flag.String(
	"address",
	"",
	"The host:port that the server is bound to.",
)

var taskHandlerAddress = flag.String(
	"taskHandlerAddress",
	"127.0.0.1:1169", // "taskhandler".each_char.collect(&:ord).inject(:+)
	"The host:port for the internal task completion callback",
)

var heartbeatInterval = flag.Duration(
	"heartbeatInterval",
	60*time.Second,
	"the interval between heartbeats for maintaining presence",
)

var etcdCluster = flag.String(
	"etcdCluster",
	"http://127.0.0.1:4001",
	"Comma-separated list of etcd addresses (http://ip:port).",
)

var corsEnabled = flag.Bool(
	"corsEnabled",
	false,
	"Enable CORS",
)

var username = flag.String(
	"username",
	"",
	"Username for basic auth, enables basic auth if set.",
)

var password = flag.String(
	"password",
	"",
	"Password for basic auth.",
)

var natsAddresses = flag.String(
	"natsAddresses",
	"",
	"Comma-separated list of NATS addresses (ip:port).",
)

var natsUsername = flag.String(
	"natsUsername",
	"",
	"Username to connect to nats.",
)

var natsPassword = flag.String(
	"natsPassword",
	"",
	"Password for nats user.",
)

const (
	dropsondeDestination = "localhost:3457"
	dropsondeOrigin      = "receptor"
)

func main() {
	flag.Parse()

	cf_debug_server.Run()

	logger := cf_lager.New("receptor")
	logger.Info("starting")

	initializeDropsonde(logger)

	if err := validateNatsArguments(); err != nil {
		logger.Error("invalid-nats-flags", err)
		os.Exit(1)
	}

	bbs := initializeReceptorBBS(logger)

	handler := handlers.New(bbs, logger, *username, *password, *corsEnabled)

	worker, enqueue := task_handler.NewTaskWorkerPool(bbs, logger)
	taskHandler := task_handler.New(enqueue, logger)

	members := grouper.Members{
		{"server", http_server.New(*serverAddress, handler)},
		{"worker", worker},
		{"task_complete_handler", http_server.New(*taskHandlerAddress, taskHandler)},
		{"heartbeater", initializeReceptorHeartbeat(*taskHandlerAddress, *heartbeatInterval, bbs, logger)},
	}

	if *registerWithRouter {
		registration := initializeServerRegistration(logger)
		natsClient := diegonats.NewClient()
		members = append(members, grouper.Member{
			Name:   "background_heartbeat",
			Runner: natbeat.NewBackgroundHeartbeat(natsClient, *natsAddresses, *natsUsername, *natsPassword, logger, registration),
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
	if *registerWithRouter {
		if *natsAddresses == "" || *serverDomainNames == "" {
			return errors.New("registerWithRouter is set, but nats addresses or domain names were left blank")
		}
	}
	return nil
}

func initializeDropsonde(logger lager.Logger) {
	err := dropsonde.Initialize(dropsondeDestination, dropsondeOrigin)
	if err != nil {
		logger.Error("failed to initialize dropsonde: %v", err)
	}
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

func initializeServerRegistration(logger lager.Logger) (registration natbeat.RegistryMessage) {
	domains := strings.Split(*serverDomainNames, ",")

	addressComponents := strings.Split(*serverAddress, ":")
	if len(addressComponents) != 2 {
		logger.Error("server-address-invalid", fmt.Errorf("%s is not a valid serverAddress", *serverAddress))
		os.Exit(1)
	}

	host, err := localip.LocalIP()
	if err != nil {
		logger.Error("local-ip-invalid", err)
		os.Exit(1)
	}

	port, err := strconv.Atoi(addressComponents[1])
	if err != nil {
		logger.Error("server-address-invalid", fmt.Errorf("%s does not have a valid port", *serverAddress))
		os.Exit(1)
	}

	return natbeat.RegistryMessage{
		URIs: domains,
		Host: host,
		Port: port,
	}
}

func initializeReceptorHeartbeat(taskHandlerAddress string, interval time.Duration, bbs Bbs.ReceptorBBS, logger lager.Logger) ifrit.Runner {
	guid, err := uuid.NewV4()
	if err != nil {
		logger.Error("failed-to-generate-guid", err)
		os.Exit(1)
	}

	presence := models.NewReceptorPresence(guid.String(), "http://"+taskHandlerAddress)
	return bbs.NewReceptorHeartbeat(presence, interval)
}
