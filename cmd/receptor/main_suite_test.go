package main_test

import (
	"bytes"
	"encoding/json"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tedsuo/rata"

	"testing"
)

func TestReceptor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Receptor Suite")
}

func doJSONRequest(reqGen *rata.RequestGenerator, name string, jsonObj interface{}) *http.Response {
	body, err := json.Marshal(jsonObj)
	Ω(err).ShouldNot(HaveOccurred())

	req, err := reqGen.CreateRequest(name, nil, bytes.NewReader(body))
	Ω(err).ShouldNot(HaveOccurred())

	client := http.Client{}
	res, err := client.Do(req)
	Ω(err).ShouldNot(HaveOccurred())

	return res
}
