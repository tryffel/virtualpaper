package e2e

import (
	"gopkg.in/h2non/baloo.v3"
	"testing"
)

const userName = "user"
const userPassword = "user"

const adminUser = "admin"
const adminPassw = "admin"

var test = baloo.New("http://localhost:8000")

func apiTest(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
}
