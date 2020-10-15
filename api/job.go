package api

import (
	"fmt"
	"net/http"
	"tryffel.net/go/virtualpaper/storage"
)

func (a *Api) GetJob(w http.ResponseWriter, req *http.Request) {
	userId, ok := getUserId(req)
	if !ok {
		respError(w, fmt.Errorf("user id not found: %v", storage.ErrInternalError))
		return
	}

	params, err := getPaging(req)
	if err != nil {
		respBadRequest(w, err.Error(), nil)
		return
	}

	jobs, err := a.db.JobStore.GetByUser(userId, params)
	if err != nil {
		respError(w, err)
		return
	}

	respResourceList(w, jobs, 100)
}
