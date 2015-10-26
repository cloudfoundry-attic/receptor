package handlers

import (
	"net/http"

	"github.com/cloudfoundry-incubator/receptor"
)

type VersionFilesLocator interface {
	GetVersionFile(string) string
}

type VersionHandler struct {
	locator VersionFilesLocator
}

func NewVersionHandler(locator VersionFilesLocator) *VersionHandler {
	return &VersionHandler{
		locator: locator,
	}
}

func (v *VersionHandler) GetVersion(w http.ResponseWriter, req *http.Request) {
	writeJSONResponse(w, http.StatusOK, receptor.VersionResponse{
		CfRelease:           v.locator.GetVersionFile("CF_RELEASE"),
		CfRoutingRelease:    v.locator.GetVersionFile("CF_ROUTING_RELEASE"),
		DiegoRelease:        v.locator.GetVersionFile("DIEGO_RELEASE"),
		GardenLinuxRelease:  v.locator.GetVersionFile("GARDEN_LINUX_RELEASE"),
		LatticeRelease:      v.locator.GetVersionFile("LATTICE_RELEASE"),
		LatticeReleaseImage: v.locator.GetVersionFile("LATTICE_RELEASE_IMAGE"),
		Ltc:                 v.locator.GetVersionFile("LTC"),
		Receptor:            v.locator.GetVersionFile("RECEPTOR"),
	})
}
