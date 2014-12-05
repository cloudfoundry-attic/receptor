package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	Bbs "github.com/cloudfoundry-incubator/runtime-schema/bbs"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/pivotal-golang/lager"
)

type FreshDomainHandler struct {
	bbs    Bbs.ReceptorBBS
	logger lager.Logger
}

func NewFreshDomainHandler(bbs Bbs.ReceptorBBS, logger lager.Logger) *FreshDomainHandler {
	return &FreshDomainHandler{
		bbs:    bbs,
		logger: logger.Session("fresh-domain-handler"),
	}
}

func (h *FreshDomainHandler) Bump(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("bump")
	freshDomainRequest := receptor.FreshDomainBumpRequest{}

	err := json.NewDecoder(req.Body).Decode(&freshDomainRequest)
	if err != nil {
		logger.Error("invalid-json", err)
		writeBadRequestResponse(w, receptor.InvalidJSON, err)
		return
	}

	freshness := serialization.FreshnessFromRequest(freshDomainRequest)

	err = h.bbs.BumpFreshness(freshness)
	if err != nil {
		if _, ok := err.(models.ValidationError); ok {
			logger.Error("freshness-invalid", err)
			writeBadRequestResponse(w, receptor.InvalidFreshness, err)
			return
		}

		logger.Error("bump-freshness-failed", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *FreshDomainHandler) GetAll(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("get-all")

	freshnesses, err := h.bbs.Freshnesses()
	if err != nil {
		logger.Error("failed-to-fetch-freshnesses", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	responses := make([]receptor.FreshDomainResponse, 0, len(freshnesses))
	for _, freshness := range freshnesses {
		responses = append(responses, serialization.FreshnessToResponse(freshness))
	}

	writeJSONResponse(w, http.StatusOK, responses)
}
