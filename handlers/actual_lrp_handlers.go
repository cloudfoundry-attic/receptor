package handlers

import (
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
