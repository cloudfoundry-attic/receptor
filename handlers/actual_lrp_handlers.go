package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	Bbs "github.com/cloudfoundry-incubator/runtime-schema/bbs"
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
		logger: logger,
	}
}

func (h *ActualLRPHandler) GetAll(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("get-all-actual-lrps-handler")

	actualLRPs, err := h.bbs.ActualLRPs()
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

	actualLRPs, err := h.bbs.ActualLRPsByDomain(domain)
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

	indexString := req.FormValue("index")

	var actualLRPs []models.ActualLRP
	var err error

	if indexString == "" {
		actualLRPs, err = h.bbs.ActualLRPsByProcessGuid(processGuid)
	} else {
		logger = logger.Session("and-index", lager.Data{
			"Index": indexString,
		})

		index, indexErr := strconv.Atoi(indexString)
		if indexErr != nil {
			err = errors.New("index not a number")
			logger.Error("invalid-index", err)
			writeBadRequestResponse(w, receptor.InvalidRequest, err)
			return
		}

		actualLRPs, err = h.bbs.ActualLRPsByProcessGuidAndIndex(processGuid, index)
	}

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

func (h *ActualLRPHandler) KillByProcessGuidAndIndex(w http.ResponseWriter, req *http.Request) {
	processGuid := req.FormValue(":process_guid")
	indexString := req.FormValue("index")
	logger := h.logger.Session("kill-by-process-guid-and-index-actual-lrps-handler", lager.Data{
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

	actualLRPs, err := h.bbs.ActualLRPsByProcessGuidAndIndex(processGuid, index)
	if err != nil {
		logger.Error("failed-to-fetch-actual-lrps-by-process-guid-and-index", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	if len(actualLRPs) == 0 {
		errorMessage := fmt.Sprintf("process-guid '%s' does not exist or has no instances at index %d", processGuid, index)
		logger.Error("no-instances-to-delete", errors.New(errorMessage))
		writeJSONResponse(w, http.StatusNotFound, receptor.Error{
			Type:    receptor.ActualLRPIndexNotFound,
			Message: errorMessage,
		})
		return
	}

	stopInstances := make([]models.StopLRPInstance, 0, len(actualLRPs))

	for _, actualLRP := range actualLRPs {
		stopInstances = append(stopInstances, models.StopLRPInstance{
			ProcessGuid:  actualLRP.ProcessGuid,
			Index:        actualLRP.Index,
			InstanceGuid: actualLRP.InstanceGuid,
		})
	}

	err = h.bbs.RequestStopLRPInstances(stopInstances)
	if err != nil {
		logger.Error("failed-to-request-stop-lrp-instances", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
