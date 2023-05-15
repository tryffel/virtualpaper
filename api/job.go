package api

import (
	"github.com/labstack/echo/v4"
)

func (a *Api) GetJob(c echo.Context) error {
	ctx := c.(UserContext)
	params, err := bindPaging(c)
	if err != nil {
		return err
	}

	jobs, err := a.db.JobStore.GetJobsByUserId(ctx.UserId, params)
	if err != nil {
		return err
	}

	return resourceList(c, jobs, len(*jobs))
}
