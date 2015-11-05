package handlers

import (
	"net/http"

	"github.com/cloudfoundry-incubator/bbs"
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	"github.com/pivotal-golang/lager"
)

type CellHandler struct {
	serviceClient bbs.ServiceClient
	logger        lager.Logger
}

func NewCellHandler(serviceClient bbs.ServiceClient, logger lager.Logger) *CellHandler {
	return &CellHandler{
		serviceClient: serviceClient,
		logger:        logger.Session("cell-handler"),
	}
}

func (h *CellHandler) GetAll(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("get-all")

	cellPresences, err := h.serviceClient.Cells(logger)
	if err != nil {
		logger.Error("failed-to-fetch-cells", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	responses := make([]receptor.CellResponse, 0, len(cellPresences))
	for _, cellPresence := range cellPresences {
		responses = append(responses, serialization.CellPresenceToCellResponse(*cellPresence))
	}

	writeJSONResponse(w, http.StatusOK, responses)
}
