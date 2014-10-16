package testrunner

import (
	"os/exec"

	"github.com/tedsuo/ifrit/ginkgomon"
)

func New(binPath string, address, etcdUrl string) *ginkgomon.Runner {
	return ginkgomon.New(ginkgomon.Config{
		Name: "receptor",
		Command: exec.Command(binPath,
			"-address", address,
			"-etcdCluster", etcdUrl,
		),
		StartCheck: "started",
	})
}
