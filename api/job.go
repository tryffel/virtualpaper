package api

import (
	"fmt"
	"net/http"
	"tryffel.net/go/virtualpaper/errors"
)

func (a *Api) GetJob(w http.ResponseWriter, req *http.Request) {
	handler := "api.GetJob"
	userId, ok := getUserId(req)
	if !ok {
		respError(w, fmt.Errorf("user id not found: %v", errors.ErrInternalError), handler)
		return
	}

	params, err := getPaging(req)
	if err != nil {
		respBadRequest(w, err.Error(), nil)
		return
	}

	jobs, err := a.db.JobStore.GetByUser(userId, params)
	if err != nil {
		respError(w, err, handler)
		return
	}

	respResourceList(w, jobs, 100)
}
