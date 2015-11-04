package handlers

import (
	"net/http"

	"github.com/cloudfoundry-incubator/locket"
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	"github.com/pivotal-golang/lager"
)

type CellHandler struct {
	locketClient locket.Client
	logger       lager.Logger
}

func NewCellHandler(locketClient locket.Client, logger lager.Logger) *CellHandler {
	return &CellHandler{
		locketClient: locketClient,
		logger:       logger.Session("cell-handler"),
	}
}

func (h *CellHandler) GetAll(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("get-all")

	cellPresences, err := h.locketClient.Cells()
	if err != nil {
		logger.Error("failed-to-fetch-cells", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	responses := make([]receptor.CellResponse, 0, len(cellPresences))
	for _, cellPresence := range cellPresences {
		responses = append(responses, serialization.CellPresenceToCellResponse(cellPresence))
	}

	writeJSONResponse(w, http.StatusOK, responses)
}
