package task_watcher_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestTaskWatcher(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TaskWatcher Suite")
}
