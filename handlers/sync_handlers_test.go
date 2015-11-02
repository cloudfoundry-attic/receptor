package handlers_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/cloudfoundry-incubator/receptor/handlers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager"
)

var _ = Describe("Sync Handlers", func() {
	var (
		logger           lager.Logger
		responseRecorder *httptest.ResponseRecorder
		fakeLocator      *fakeArtifactLocator
		handler          *handlers.SyncHandler
	)

	BeforeEach(func() {
		logger = lager.NewLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
		responseRecorder = httptest.NewRecorder()
		fakeLocator = &fakeArtifactLocator{}
		handler = handlers.NewSyncHandler(fakeLocator, logger)
	})

	Describe("Download", func() {
		JustBeforeEach(func() {
			req := newTestRequest("")
			req.Form = url.Values{":arch": []string{"windows"}, ":artifact": []string{"ltc.exe"}}
			handler.Download(responseRecorder, req)
		})

		Context("when the artifact exists", func() {
			BeforeEach(func() {
				fakeLocator.buffer = bytes.NewReader([]byte("xyz"))
			})

			It("sends the contents", func() {
				Expect(fakeLocator.paramArch).To(Equal("windows"))
				Expect(fakeLocator.paramName).To(Equal("ltc.exe"))

				Expect(responseRecorder.Body.Bytes()).To(Equal([]byte("xyz")))
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				Expect(responseRecorder.HeaderMap.Get("Content-length")).To(Equal("3"))
				Expect(responseRecorder.HeaderMap.Get("Content-type")).To(Equal("application/octet-stream"))
			})
		})

		Context("when the artifact locator fails", func() {
			BeforeEach(func() {
				fakeLocator.err = errors.New("failed")
			})

			It("returns an error", func() {
				Expect(fakeLocator.paramArch).To(Equal("windows"))
				Expect(fakeLocator.paramName).To(Equal("ltc.exe"))

				Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
			})
		})

	})
})

type fakeArtifactLocator struct {
	buffer *bytes.Reader
	err    error

	paramArch, paramName string
}

func (l *fakeArtifactLocator) LocateArtifact(arch, name string) (io.ReadSeeker, error) {
	l.paramArch = arch
	l.paramName = name
	return l.buffer, l.err
}
