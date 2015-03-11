package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	Bbs "github.com/cloudfoundry-incubator/runtime-schema/bbs"
	"github.com/cloudfoundry-incubator/runtime-schema/bbs/bbserrors"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/pivotal-golang/lager"
)

type VolumeHandler struct {
	bbs    Bbs.ReceptorBBS
	logger lager.Logger
}

func NewVolumeHandler(bbs Bbs.ReceptorBBS, logger lager.Logger) *VolumeHandler {
	return &VolumeHandler{
		bbs:    bbs,
		logger: logger.Session("task-handler"),
	}
}

func (h *VolumeHandler) CreateVolumeSet(w http.ResponseWriter, r *http.Request) {
	log := h.logger.Session("create-volume-set")
	volSetRequest := receptor.VolumeSetCreateRequest{}

	err := json.NewDecoder(r.Body).Decode(&volSetRequest)
	if err != nil {
		log.Error("invalid-json", err)
		writeJSONResponse(w, http.StatusBadRequest, receptor.Error{
			Type:    receptor.InvalidJSON,
			Message: err.Error(),
		})
		return
	}

	volSet := serialization.VolumeSetFromCreateRequest(volSetRequest)

	log.Debug("creating-volume-set", lager.Data{"volume-set-guid": volSet.VolumeSetGuid})

	err = h.bbs.DesireVolumeSet(log, volSet)
	if err != nil {
		log.Error("failed-to-desire-volume-set", err)

		if _, ok := err.(models.ValidationError); ok {
			writeJSONResponse(w, http.StatusBadRequest, receptor.Error{
				Type:    receptor.InvalidVolumeSet,
				Message: err.Error(),
			})
			return
		}

		if err == bbserrors.ErrStoreResourceExists {
			writeJSONResponse(w, http.StatusConflict, receptor.Error{
				Type:    receptor.VolumeSetGuidAlreadyExists,
				Message: "volume set already exists",
			})
		} else {
			writeUnknownErrorResponse(w, err)
		}
		return
	}

	log.Info("created", lager.Data{"volume-set-guid": volSet.VolumeSetGuid})
	w.WriteHeader(http.StatusCreated)
}

func (h *VolumeHandler) VolumesByVolumeSetGuid(w http.ResponseWriter, req *http.Request) {
	guid := req.FormValue(":volume_set_guid")
	logger := h.logger.Session("get-volumes-by-volume-set-guid", lager.Data{
		"volume-set-guid": guid,
	})

	var vols []models.Volume
	var err error

	vols, err = h.bbs.VolumesByVolumeSetGuid(logger, guid)

	writeVolumeResponse(w, logger, vols, err)
}

func writeVolumeResponse(w http.ResponseWriter, logger lager.Logger, vols []models.Volume, err error) {
	if err != nil {
		logger.Error("failed-to-fetch-volumes", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	volumeResponses := []receptor.VolumeResponse{}
	for i := range vols {
		volumeResponses = append(volumeResponses, serialization.VolumeToResponse(vols[i]))
	}

	writeJSONResponse(w, http.StatusOK, volumeResponses)
}
