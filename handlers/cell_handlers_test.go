package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/cloudfoundry-incubator/bbs/fake_bbs"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/handlers"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager"
)

var _ = Describe("Cell Handlers", func() {
	var (
		logger           lager.Logger
		serviceClient    *fake_bbs.FakeServiceClient
		responseRecorder *httptest.ResponseRecorder
		handler          *handlers.CellHandler
	)

	BeforeEach(func() {
		serviceClient = new(fake_bbs.FakeServiceClient)
		logger = lager.NewLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
		responseRecorder = httptest.NewRecorder()
		handler = handlers.NewCellHandler(serviceClient, logger)
	})

	Describe("GetAll", func() {
		var cellPresences models.CellSet

		BeforeEach(func() {
			capacity := models.NewCellCapacity(128, 1024, 6)
			cellPresences = models.CellSet{}
			cellPresence0 := models.NewCellPresence("cell-id-0", "1.2.3.4", "the-zone", capacity, []string{"provider-0"}, []string{"stack-0"})
			cellPresences.Add(&cellPresence0)
			cellPresence1 := models.NewCellPresence("cell-id-1", "4.5.6.7", "the-zone", capacity, []string{"provider-1"}, []string{"stack-1"})
			cellPresences.Add(&cellPresence1)
		})

		JustBeforeEach(func() {
			handler.GetAll(responseRecorder, newTestRequest(""))
		})

		Context("when reading Cells from BBS succeeds", func() {
			BeforeEach(func() {
				serviceClient.CellsReturns(cellPresences, nil)
			})

			It("call the BBS to retrieve the actual LRPs", func() {
				Expect(serviceClient.CellsCallCount()).To(Equal(1))
			})

			It("responds with 200 Status OK", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})

			It("returns a list of cell presence responses", func() {
				response := []receptor.CellResponse{}
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())

				Expect(response).To(HaveLen(2))
				for _, cellPresence := range cellPresences {
					Expect(response).To(ContainElement(serialization.CellPresenceToCellResponse(*cellPresence)))
				}
			})
		})

		Context("when the BBS returns no cells", func() {
			BeforeEach(func() {
				serviceClient.CellsReturns(models.CellSet{}, nil)
			})

			It("responds with 200 Status OK", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})

			It("returns an empty list", func() {
				Expect(responseRecorder.Body.String()).To(Equal("[]"))
			})
		})

		Context("when reading from the BBS fails", func() {
			BeforeEach(func() {
				serviceClient.CellsReturns(models.CellSet{}, errors.New("Something went wrong"))
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
			})

			It("provides relevant error information", func() {
				var receptorError receptor.Error
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &receptorError)
				Expect(err).NotTo(HaveOccurred())

				Expect(receptorError).To(Equal(receptor.Error{
					Type:    receptor.UnknownError,
					Message: "Something went wrong",
				}))

			})
		})
	})
})
