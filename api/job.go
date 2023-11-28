package api

import (
	"github.com/labstack/echo/v4"
)

func (a *Api) GetJob(c echo.Context) error {
	ctx := c.(UserContext)
	pagination := getPagination(c)
	jobs, err := a.db.JobStore.GetJobsByUserId(ctx.UserId, pagination.toPagination())
	if err != nil {
		return err
	}

	return resourceList(c, jobs, len(*jobs))
}
