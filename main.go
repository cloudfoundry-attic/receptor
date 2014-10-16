package main

import (
	"flag"
	"os"

	"github.com/cloudfoundry-incubator/cf-debug-server"
	"github.com/cloudfoundry-incubator/cf-lager"
	"github.com/cloudfoundry-incubator/receptor/server"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/sigmon"
)

var serverAddress = flag.String(
	"address",
	"",
	"Specifies the address to bind to",
)

func main() {
	var err error

	flag.Parse()

	logger := cf_lager.New("receptor")

	logger.Info("starting")

	server := &server.Server{
		Address: *serverAddress,
		Logger:  logger,
	}

	group := grouper.NewOrdered(os.Interrupt, grouper.Members{
		{"server", server},
	})

	cf_debug_server.Run()

	monitor := ifrit.Invoke(sigmon.New(group))

	logger.Info("started")

	err = <-monitor.Wait()
	if err != nil {
		logger.Error("exited-with-failure", err)
		os.Exit(1)
	}

	logger.Info("exited")
}
