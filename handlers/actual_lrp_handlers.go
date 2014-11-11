package handlers

import (
	"errors"
	"net/http"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	Bbs "github.com/cloudfoundry-incubator/runtime-schema/bbs"
	"github.com/pivotal-golang/lager"
)

type ActualLRPHandler struct {
	bbs    Bbs.ReceptorBBS
	logger lager.Logger
}

func NewActualLRPHandler(bbs Bbs.ReceptorBBS, logger lager.Logger) *ActualLRPHandler {
	return &ActualLRPHandler{
		bbs:    bbs,
		logger: logger,
	}
}

func (h *ActualLRPHandler) GetAll(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("get-all-actual-lrps-handler")

	actualLRPs, err := h.bbs.GetAllActualLRPs()
	if err != nil {
		logger.Error("failed-to-fetch-actual-lrps", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	responses := make([]receptor.ActualLRPResponse, 0, len(actualLRPs))
	for _, actualLRP := range actualLRPs {
		responses = append(responses, serialization.ActualLRPToResponse(actualLRP))
	}

	writeJSONResponse(w, http.StatusOK, responses)
}

func (h *ActualLRPHandler) GetAllByDomain(w http.ResponseWriter, req *http.Request) {
	domain := req.FormValue(":domain")
	logger := h.logger.Session("get-all-by-domain-actual-lrps-handler", lager.Data{
		"Domain": domain,
	})

	if domain == "" {
		err := errors.New("domain missing from request")
		logger.Error("missing-domain", err)
		writeBadRequestResponse(w, receptor.InvalidRequest, err)
		return
	}

	actualLRPs, err := h.bbs.GetAllActualLRPsByDomain(domain)
	if err != nil {
		logger.Error("failed-to-fetch-actual-lrps-by-domain", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	responses := make([]receptor.ActualLRPResponse, 0, len(actualLRPs))
	for _, actualLRP := range actualLRPs {
		responses = append(responses, serialization.ActualLRPToResponse(actualLRP))
	}

	writeJSONResponse(w, http.StatusOK, responses)
}

func (h *ActualLRPHandler) GetAllByProcessGuid(w http.ResponseWriter, req *http.Request) {
	processGuid := req.FormValue(":process_guid")
	logger := h.logger.Session("get-all-by-process-guid-actual-lrps-handler", lager.Data{
		"ProcessGuid": processGuid,
	})

	if processGuid == "" {
		err := errors.New("process_guid missing from request")
		logger.Error("missing-process-guid", err)
		writeBadRequestResponse(w, receptor.InvalidRequest, err)
		return
	}

	actualLRPs, err := h.bbs.GetActualLRPsByProcessGuid(processGuid)
	if err != nil {
		logger.Error("failed-to-fetch-actual-lrps-by-process-guid", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	responses := make([]receptor.ActualLRPResponse, 0, len(actualLRPs))
	for _, actualLRP := range actualLRPs {
		responses = append(responses, serialization.ActualLRPToResponse(actualLRP))
	}

	writeJSONResponse(w, http.StatusOK, responses)
}
