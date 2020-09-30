package proxyHttp

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/kataras/golog"
	"io/ioutil"
	"mitm-proxy/app/models"
	"net/http"
	"strconv"
	"strings"
	"time"
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

func (h *handler) CheckXXE(w http.ResponseWriter, r *http.Request) {
	stringId := mux.Vars(r)["id"]
	id, err := strconv.ParseUint(stringId, 10, 64)
	if err != nil {
		golog.Error(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	request, err := h.useCase.GetRequest(id)
	if err != nil {
		golog.Error(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	xxeBody := h.useCase.AddXXEEntity(request.Body)

	httpRequest := models.ToHttpRequest(request)
	httpRequest.Body = ioutil.NopCloser(strings.NewReader(xxeBody))
	httpRequest.ContentLength = int64(len(xxeBody))

	client := http.Client{
		Timeout: 5 * time.Second,
	}
	response, err := client.Do(httpRequest)
	if err != nil {
		golog.Error(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		golog.Error(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	var bodyJson []byte
	type XXE struct {
		XXE bool `json:"xxe"`
	}
	if strings.Contains(string(body), "root:") {
		bodyJson, _ = json.Marshal(XXE{true})
	} else {
		bodyJson, _ = json.Marshal(XXE{false})
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(bodyJson)
}
