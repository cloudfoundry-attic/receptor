package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	Bbs "github.com/cloudfoundry-incubator/runtime-schema/bbs"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/cloudfoundry/storeadapter"
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
	desireLRPRequest := receptor.DesiredLRPCreateRequest{}

	err := json.NewDecoder(r.Body).Decode(&desireLRPRequest)
	if err != nil {
		log.Error("invalid-json", err)
		writeBadRequestResponse(w, receptor.InvalidJSON, err)
		return
	}

	desiredLRP, err := serialization.DesiredLRPFromRequest(desireLRPRequest)
	if err != nil {
		log.Error("lrp-request-invalid", err)
		writeBadRequestResponse(w, receptor.InvalidLRP, err)
		return
	}

	err = h.bbs.DesireLRP(desiredLRP)
	if err != nil {
		log.Error("desire-lrp-failed", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *DesiredLRPHandler) Get(w http.ResponseWriter, r *http.Request) {
	processGuid := r.FormValue(":process_guid")
	log := h.logger.Session("get-desired-lrp-by-process-guid-handler", lager.Data{
		"ProcessGuid": processGuid,
	})

	if processGuid == "" {
		err := errors.New("process_guid missing from request")
		log.Error("missing-process-guid", err)
		writeBadRequestResponse(w, receptor.InvalidRequest, err)
		return
	}

	desiredLRP, err := h.bbs.GetDesiredLRPByProcessGuid(processGuid)
	if err == storeadapter.ErrorKeyNotFound {
		writeLRPNotFoundResponse(w)
		return
	}

	if err != nil {
		log.Error("unknown-error", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	response := serialization.DesiredLRPToResponse(desiredLRP)

	writeJSONResponse(w, http.StatusOK, response)
}

func (h *DesiredLRPHandler) Update(w http.ResponseWriter, r *http.Request) {
	processGuid := r.FormValue(":process_guid")
	log := h.logger.Session("update-desired-lrp-handler", lager.Data{
		"ProcessGuid": processGuid,
	})

	if processGuid == "" {
		err := errors.New("process_guid missing from request")
		log.Error("missing-process-guid", err)
		writeBadRequestResponse(w, receptor.InvalidRequest, err)
		return
	}

	desireLRPRequest := receptor.DesiredLRPUpdateRequest{}

	err := json.NewDecoder(r.Body).Decode(&desireLRPRequest)
	if err != nil {
		log.Error("invalid-json", err)
		writeBadRequestResponse(w, receptor.InvalidJSON, err)
		return
	}

	update := serialization.DesiredLRPUpdateFromRequest(desireLRPRequest)

	err = h.bbs.UpdateDesiredLRP(processGuid, update)
	if err == storeadapter.ErrorKeyNotFound {
		writeLRPNotFoundResponse(w)
		return
	}

	if err != nil {
		log.Error("unknown-error", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *DesiredLRPHandler) Delete(w http.ResponseWriter, req *http.Request) {
	processGuid := req.FormValue(":process_guid")
	log := h.logger.Session("delete-desired-lrp-handler", lager.Data{
		"ProcessGuid": processGuid,
	})

	if processGuid == "" {
		err := errors.New("process_guid missing from request")
		log.Error("missing-process-guid", err)
		writeBadRequestResponse(w, receptor.InvalidRequest, err)
		return
	}

	err := h.bbs.RemoveDesiredLRPByProcessGuid(processGuid)

	if err == storeadapter.ErrorKeyNotFound {
		writeLRPNotFoundResponse(w)
		return
	}

	if err != nil {
		log.Error("unknown-error", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *DesiredLRPHandler) GetAll(w http.ResponseWriter, req *http.Request) {
	desiredLRPs, err := h.bbs.GetAllDesiredLRPs()
	writeDesiredLRPResponse(w, h.logger.Session("get-all-desired-lrps-handler"), desiredLRPs, err)
}

func (h *DesiredLRPHandler) GetAllByDomain(w http.ResponseWriter, req *http.Request) {
	log := h.logger.Session("get-all-desired-lrps-by-domain-handler")

	lrpDomain := req.FormValue(":domain")
	if lrpDomain == "" {
		err := errors.New("domain missing from request")
		log.Error("missing-domain", err)
		writeBadRequestResponse(w, receptor.InvalidRequest, err)
		return
	}

	desiredLRPs, err := h.bbs.GetAllDesiredLRPsByDomain(lrpDomain)
	writeDesiredLRPResponse(w, log, desiredLRPs, err)
}

func writeDesiredLRPResponse(w http.ResponseWriter, logger lager.Logger, desiredLRPs []models.DesiredLRP, err error) {
	if err != nil {
		logger.Error("failed-to-fetch-desired-lrps", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	responses := make([]receptor.DesiredLRPResponse, 0, len(desiredLRPs))
	for _, desiredLRP := range desiredLRPs {
		responses = append(responses, serialization.DesiredLRPToResponse(desiredLRP))
	}

	writeJSONResponse(w, http.StatusOK, responses)
}

func writeLRPNotFoundResponse(w http.ResponseWriter) {
	writeJSONResponse(w, http.StatusNotFound, receptor.Error{
		Type:    receptor.LRPNotFound,
		Message: "LRP not found",
	})
}
