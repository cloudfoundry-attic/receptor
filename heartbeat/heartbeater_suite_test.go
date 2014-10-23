package heartbeat_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestHeartbeater(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Heartbeat Suite")
}
