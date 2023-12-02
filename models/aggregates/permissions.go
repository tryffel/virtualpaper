package aggregates

import "tryffel.net/go/virtualpaper/models"

type DocumentPermissions struct {
	UserId            int
	Document          string
	Owner             bool
	SharedPermissions models.Permissions
}
