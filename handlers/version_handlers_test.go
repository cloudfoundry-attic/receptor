package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/handlers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Version Handlers", func() {
	var (
		responseRecorder *httptest.ResponseRecorder
		fakeLocator      *fakeVersionFilesLocator
		handler          *handlers.VersionHandler
	)

	BeforeEach(func() {
		responseRecorder = httptest.NewRecorder()
		fakeLocator = &fakeVersionFilesLocator{}
		handler = handlers.NewVersionHandler(fakeLocator)
	})

	Describe("GetVersion", func() {
		Context("when version files exist", func() {
			BeforeEach(func() {
				fakeLocator.cfRelease = "v218"
				fakeLocator.cfRoutingRelease = "v219"
				fakeLocator.diegoRelease = "v220"
				fakeLocator.gardenLinuxRelease = "v221"
				fakeLocator.latticeRelease = "v222"
				fakeLocator.latticeReleaseImage = "v223"
				fakeLocator.ltc = "v224"
				fakeLocator.receptor = "v225"
			})

			It("returns the versions", func() {
				req := newTestRequest("")
				handler.GetVersion(responseRecorder, req)

				Expect(responseRecorder.Code).To(Equal(http.StatusOK))

				response := receptor.VersionResponse{}
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response.CfRelease).To(Equal("v218"))
				Expect(response.CfRoutingRelease).To(Equal("v219"))
				Expect(response.DiegoRelease).To(Equal("v220"))
				Expect(response.GardenLinuxRelease).To(Equal("v221"))
				Expect(response.LatticeRelease).To(Equal("v222"))
				Expect(response.LatticeReleaseImage).To(Equal("v223"))
				Expect(response.Ltc).To(Equal("v224"))
				Expect(response.Receptor).To(Equal("v225"))
			})
		})

		Context("when no version files exist", func() {
			It("returns an empty hash", func() {
				req := newTestRequest("")
				handler.GetVersion(responseRecorder, req)

				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				Expect(responseRecorder.Body.String()).To(Equal("{}"))
			})
		})
	})
})

type fakeVersionFilesLocator struct {
	cfRelease, cfRoutingRelease, diegoRelease, gardenLinuxRelease, latticeRelease, latticeReleaseImage, ltc, receptor string
}

func (v *fakeVersionFilesLocator) GetVersionFile(filename string) string {
	switch filename {
	case "CF_RELEASE":
		return v.cfRelease
	case "CF_ROUTING_RELEASE":
		return v.cfRoutingRelease
	case "DIEGO_RELEASE":
		return v.diegoRelease
	case "GARDEN_LINUX_RELEASE":
		return v.gardenLinuxRelease
	case "LATTICE_RELEASE":
		return v.latticeRelease
	case "LATTICE_RELEASE_IMAGE":
		return v.latticeReleaseImage
	case "LTC":
		return v.ltc
	case "RECEPTOR":
		return v.receptor
	default:
		Fail("unknown version file " + filename)
	}

	return ""
}
