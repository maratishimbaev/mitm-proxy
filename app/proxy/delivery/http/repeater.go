package proxyHttp

import (
	"encoding/json"
	"net/http"
)

func (h *handler) GetRequests(w http.ResponseWriter, r *http.Request) {
	requests, err := h.useCase.GetRequests()
	if err != nil {
		status := http.StatusInternalServerError
		http.Error(w, http.StatusText(status), status)
		return
	}

	requestsJson, err := json.Marshal(requests)
	if err != nil {
		status := http.StatusInternalServerError
		http.Error(w, http.StatusText(status), status)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(requestsJson)
}
