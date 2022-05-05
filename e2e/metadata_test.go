package e2e

import "testing"

func TestMetadata(t *testing.T) {
	apiTest(t)

	TestLogin(t)

	addMetadata(t)
}
