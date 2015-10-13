package handlers

import (
	"io"
	"net/http"
	"time"

	"github.com/pivotal-golang/lager"
)

type ArtifactLocator interface {
	LocateArtifact(arch, name string) (io.ReadSeeker, error)
}

type SyncHandler struct {
	locator ArtifactLocator
	logger  lager.Logger
}

func NewSyncHandler(locator ArtifactLocator, logger lager.Logger) *SyncHandler {
	return &SyncHandler{
		locator: locator,
		logger:  logger.Session("sync-handler"),
	}
}

func (h *SyncHandler) Download(w http.ResponseWriter, req *http.Request) {
	arch := req.FormValue(":arch")
	logger := h.logger.Session("download", lager.Data{
		"arch": arch,
	})

	artifactSeeker, err := h.locator.LocateArtifact(arch, "ltc")
	if err != nil {
		logger.Error("failed-to-locate-artifact", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeContent(w, req, "ltc", time.Time{}, artifactSeeker)
}
