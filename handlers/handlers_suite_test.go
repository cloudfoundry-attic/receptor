package handlers_test

import (
	"bytes"
	"io"
	"net/http"

	"github.com/cloudfoundry-incubator/receptor"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestHandlers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Handlers Suite")
}

func newTestRequest(body interface{}) *http.Request {
	var reader io.Reader
	switch body := body.(type) {
	case receptor.JSONReader:
		reader = body.JSONReader()
	case string:
		reader = bytes.NewBufferString(body)
	case []byte:
		reader = bytes.NewBuffer(body)
	}
	request, err := http.NewRequest("", "", reader)
	Ω(err).ToNot(HaveOccurred())
	return request
}
