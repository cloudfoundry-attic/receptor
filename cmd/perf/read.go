package main

import (
	"github.com/cloudfoundry-incubator/runtime-schema/bbs/shared"

	"github.com/cloudfoundry/gunk/workpool"
	"github.com/cloudfoundry/storeadapter/etcdstoreadapter"
	"github.com/pivotal-golang/lager"
)

func main() {
	logger := lager.NewLogger("perf")
	workPool, err := workpool.NewWorkPool(10)
	if err != nil {
		logger.Fatal("failed-to-construct-etcd-adapter-workpool", err, lager.Data{"num-workers": 100}) // should never happen
	}

	options := &etcdstoreadapter.ETCDOptions{
		ClusterUrls: []string{"http://127.0.0.1:4006"},
	}

	etcdAdapter, err := etcdstoreadapter.New(options, workPool)

	_, err = etcdAdapter.ListRecursively(shared.DesiredLRPSchemaRoot)
	if err != nil {
		logger.Fatal("list", err)
	}
}
