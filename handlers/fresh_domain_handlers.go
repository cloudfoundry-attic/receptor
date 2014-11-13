package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	Bbs "github.com/cloudfoundry-incubator/runtime-schema/bbs"
	"github.com/pivotal-golang/lager"
)

type FreshDomainHandler struct {
	bbs    Bbs.ReceptorBBS
	logger lager.Logger
}

func NewFreshDomainHandler(bbs Bbs.ReceptorBBS, logger lager.Logger) *FreshDomainHandler {
	return &FreshDomainHandler{
		bbs:    bbs,
		logger: logger,
	}
}

func (h *FreshDomainHandler) Create(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("create-fresh-domain-handler")
	freshDomainRequest := receptor.FreshDomainCreateRequest{}

	err := json.NewDecoder(req.Body).Decode(&freshDomainRequest)
	if err != nil {
		logger.Error("invalid-json", err)
		writeBadRequestResponse(w, receptor.InvalidJSON, err)
		return
	}

	freshness := serialization.FreshnessFromRequest(freshDomainRequest)

	err = freshness.Validate()
	if err != nil {
		logger.Error("freshness-invalid", err)
		writeBadRequestResponse(w, receptor.InvalidFreshness, err)
		return
	}

	err = h.bbs.BumpFreshness(freshness)
	if err != nil {
		logger.Error("bump-freshness-failed", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
