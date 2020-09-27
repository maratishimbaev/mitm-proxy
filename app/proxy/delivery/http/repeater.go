package proxyHttp

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"mitm-proxy/app/models"
	"net/http"
	"strconv"
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

func (h *handler) RepeatRequest(w http.ResponseWriter, r *http.Request) {
	stringId := mux.Vars(r)["id"]
	id, err := strconv.ParseUint(stringId, 10, 64)
	if err != nil {
		status := http.StatusInternalServerError
		http.Error(w, http.StatusText(status), status)
		return
	}

	request, err := h.useCase.GetRequest(id)
	if err != nil {
		status := http.StatusInternalServerError
		http.Error(w, http.StatusText(status), status)
		return
	}

	httpRequest := models.ToHttpRequest(request)
	h.ServeHTTP(w, httpRequest)
}
