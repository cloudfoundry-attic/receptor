package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	Bbs "github.com/cloudfoundry-incubator/runtime-schema/bbs"
	"github.com/cloudfoundry-incubator/runtime-schema/bbs/bbserrors"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/pivotal-golang/lager"
)

type ActualLRPHandler struct {
	bbs    Bbs.ReceptorBBS
	logger lager.Logger
}

func NewActualLRPHandler(bbs Bbs.ReceptorBBS, logger lager.Logger) *ActualLRPHandler {
	return &ActualLRPHandler{
		bbs:    bbs,
		logger: logger.Session("actual-lrp-handler"),
	}
}

func (h *ActualLRPHandler) GetAll(w http.ResponseWriter, req *http.Request) {
	domain := req.FormValue("domain")
	logger := h.logger.Session("get-all", lager.Data{
		"domain": domain,
	})

	var actualLRPs []models.ActualLRP
	var err error

	if domain == "" {
		actualLRPs, err = h.bbs.ActualLRPs()
	} else {
		actualLRPs, err = h.bbs.ActualLRPsByDomain(domain)
	}

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

func (h *ActualLRPHandler) GetAllByProcessGuid(w http.ResponseWriter, req *http.Request) {
	processGuid := req.FormValue(":process_guid")
	logger := h.logger.Session("get-all-by-process-guid", lager.Data{
		"ProcessGuid": processGuid,
	})

	if processGuid == "" {
		err := errors.New("process_guid missing from request")
		logger.Error("missing-process-guid", err)
		writeBadRequestResponse(w, receptor.InvalidRequest, err)
		return
	}

	actualLRPs, err := h.bbs.ActualLRPsByProcessGuid(processGuid)
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

func (h *ActualLRPHandler) GetByProcessGuidAndIndex(w http.ResponseWriter, req *http.Request) {
	processGuid := req.FormValue(":process_guid")
	indexString := req.FormValue(":index")

	logger := h.logger.Session("get-by-process-guid-and-index", lager.Data{
		"ProcessGuid": processGuid,
		"Index":       indexString,
	})

	if processGuid == "" {
		err := errors.New("process_guid missing from request")
		logger.Error("missing-process-guid", err)
		writeBadRequestResponse(w, receptor.InvalidRequest, err)
		return
	}

	if indexString == "" {
		err := errors.New("index missing from request")
		logger.Error("missing-index", err)
		writeBadRequestResponse(w, receptor.InvalidRequest, err)
		return
	}

	var err error

	index, indexErr := strconv.Atoi(indexString)
	if indexErr != nil {
		err = errors.New("index not a number")
		logger.Error("invalid-index", err)
		writeBadRequestResponse(w, receptor.InvalidRequest, err)
		return
	}

	actualLRP, err := h.bbs.ActualLRPByProcessGuidAndIndex(processGuid, index)
	if err == bbserrors.ErrStoreResourceNotFound {
		writeJSONResponse(w, http.StatusNotFound, nil)
		return
	}

	if err != nil {
		logger.Error("failed-to-fetch-actual-lrps-by-process-guid", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	writeJSONResponse(w, http.StatusOK, serialization.ActualLRPToResponse(*actualLRP))
}

func (h *ActualLRPHandler) KillByProcessGuidAndIndex(w http.ResponseWriter, req *http.Request) {
	processGuid := req.FormValue(":process_guid")
	indexString := req.FormValue(":index")
	logger := h.logger.Session("kill-by-process-guid-and-index", lager.Data{
		"ProcessGuid": processGuid,
		"Index":       indexString,
	})

	if processGuid == "" {
		err := errors.New("process_guid missing from request")
		logger.Error("missing-process-guid", err)
		writeBadRequestResponse(w, receptor.InvalidRequest, err)
		return
	}

	if indexString == "" {
		err := errors.New("index missing from request")
		logger.Error("missing-index", err)
		writeBadRequestResponse(w, receptor.InvalidRequest, err)
		return
	}

	index, err := strconv.Atoi(indexString)
	if err != nil {
		err = errors.New("index not a number")
		logger.Error("invalid-index", err)
		writeBadRequestResponse(w, receptor.InvalidRequest, err)
		return
	}

	actualLRP, err := h.bbs.ActualLRPByProcessGuidAndIndex(processGuid, index)
	if err != nil {
		logger.Error("failed-to-fetch-actual-lrp-by-process-guid-and-index", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	if actualLRP == nil {
		err := fmt.Errorf("process-guid '%s' does not exist or has no instance at index %d", processGuid, index)
		logger.Error("no-instances-to-delete", err)
		writeJSONResponse(w, http.StatusNotFound, receptor.Error{
			Type:    receptor.ActualLRPIndexNotFound,
			Message: err.Error(),
		})
		return
	}

	err = h.bbs.RequestStopLRPInstance(*actualLRP)
	if err != nil {
		logger.Error("failed-to-request-stop-lrp-instance", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
