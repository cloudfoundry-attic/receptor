package handlers

import (
	"net/http"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	Bbs "github.com/cloudfoundry-incubator/runtime-schema/bbs"
	"github.com/pivotal-golang/lager"
)

type CellHandler struct {
	bbs    Bbs.ReceptorBBS
	logger lager.Logger
}

func NewCellHandler(bbs Bbs.ReceptorBBS, logger lager.Logger) *CellHandler {
	return &CellHandler{
		bbs:    bbs,
		logger: logger,
	}
}

func (h *CellHandler) GetAll(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("get-all-cells-handler")

	executorPresences, err := h.bbs.GetAllExecutors()
	if err != nil {
		logger.Error("failed-to-fetch-executors", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	responses := make([]receptor.CellResponse, 0, len(executorPresences))
	for _, executorPresence := range executorPresences {
		responses = append(responses, serialization.ExecutorPresenceToCellResponse(executorPresence))
	}

	writeJSONResponse(w, http.StatusOK, responses)
}
