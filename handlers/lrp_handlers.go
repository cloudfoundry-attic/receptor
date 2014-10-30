package handlers

import (
	"encoding/json"
	"net/http"
)
import (
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	Bbs "github.com/cloudfoundry-incubator/runtime-schema/bbs"
	"github.com/pivotal-golang/lager"
)

type DesiredLRPHandler struct {
	bbs    Bbs.ReceptorBBS
	logger lager.Logger
}

func NewDesiredLRPHandler(bbs Bbs.ReceptorBBS, logger lager.Logger) *DesiredLRPHandler {
	return &DesiredLRPHandler{
		bbs:    bbs,
		logger: logger,
	}
}

func (h *DesiredLRPHandler) Create(w http.ResponseWriter, r *http.Request) {
	log := h.logger.Session("create-desired-lrp-handler")
	desireLRPRequest := receptor.CreateDesiredLRPRequest{}

	err := json.NewDecoder(r.Body).Decode(&desireLRPRequest)
	if err != nil {
		log.Error("invalid-json", err)
		writeJSONResponse(w, http.StatusBadRequest, receptor.Error{
			Type:    receptor.InvalidJSON,
			Message: err.Error(),
		})
		return
	}

	desiredLRP, err := serialization.DesiredLRPFromRequest(desireLRPRequest)
	if err != nil {
		log.Error("lrp-request-invalid", err)
		writeJSONResponse(w, http.StatusBadRequest, receptor.Error{
			Type:    receptor.InvalidLRP,
			Message: err.Error(),
		})
		return
	}

	err = h.bbs.DesireLRP(desiredLRP)
	if err != nil {
		writeJSONResponse(w, http.StatusInternalServerError, receptor.Error{
			Type:    receptor.UnknownError,
			Message: err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusCreated)
}
