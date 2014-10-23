package testrunner

import (
	"os/exec"

	"github.com/tedsuo/ifrit/ginkgomon"
)

type Args struct {
	DomainNames              string
	Address                  string
	EtcdCluster              string
	Username                 string
	Password                 string
	NatsAddresses            string
	NatsUsername             string
	NatsPassword             string
	InitialHeartbeatInterval string
}

func (args Args) ArgSlice() []string {
	return []string{
		"-domainNames", args.DomainNames,
		"-address", args.Address,
		"-etcdCluster", args.EtcdCluster,
		"-username", args.Username,
		"-password", args.Password,
		"-natsAddresses", args.NatsAddresses,
		"-natsUsername", args.NatsUsername,
		"-natsPassword", args.NatsPassword,
		"initialHeartbeatInterval", args.InitialHeartbeatInterval,
	}
}

func New(binPath string, args Args) *ginkgomon.Runner {
	return ginkgomon.New(ginkgomon.Config{
		Name:       "receptor",
		Command:    exec.Command(binPath, args.ArgSlice()...),
		StartCheck: "started",
	})
}
