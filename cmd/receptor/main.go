package main

import (
	"flag"
	"os"
	"strings"

	"github.com/cloudfoundry-incubator/cf-debug-server"
	"github.com/cloudfoundry-incubator/cf-lager"
	"github.com/cloudfoundry-incubator/receptor/handlers"
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

var serverAddress = flag.String(
	"address",
	"",
	"Specifies the address to bind to",
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

func main() {
	flag.Parse()

	cf_debug_server.Run()

	logger := cf_lager.New("receptor")
	logger.Info("starting")

	bbs := initializeReceptorBBS(logger)

	handler := handlers.New(bbs, logger, *username, *password)

	group := grouper.NewOrdered(os.Interrupt, grouper.Members{
		{"server", http_server.New(*serverAddress, handler)},
		{"watcher", task_watcher.New(bbs, logger)},
	})

	monitor := ifrit.Invoke(sigmon.New(group))

	logger.Info("started")

	err := <-monitor.Wait()
	if err != nil {
		logger.Error("exited-with-failure", err)
		os.Exit(1)
	}

	logger.Info("exited")
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
