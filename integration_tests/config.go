package integrationtest

import (
	"gopkg.in/h2non/baloo.v3"
	"os"
	"strings"
)

// loaded from env keys during startup
var serverUrl = ""
var dbHost = ""
var meiliHost = ""

type httpTest struct {
	client *baloo.Client
}

func (t *httpTest) Authorize() *httpTest {
	return &httpTest{
		client: t.client.SetHeader("Authorization", "Bearer "+UserToken),
	}
}

func (t *httpTest) AuthorizeAdmin() *httpTest {
	return &httpTest{
		client: t.client.SetHeader("Authorization", "Bearer "+AdminToken),
	}
}

func (t *httpTest) IsJson() *httpTest {
	return &httpTest{
		client: t.client.SetHeader("Content-Type", "application/json"),
	}
}

var client = &httpTest{client: baloo.New(serverUrl)}

func initConfig() {
	serverUrl = getEnv("SERVER_URL", "http://localhost:8000")
	dbHost = getEnv("DATABASE_HOST", "localhost")
	meiliHost = getEnv("MEILISEARCH_URL", "http://localhost:7700")
	client = &httpTest{client: baloo.New(serverUrl)}
}

func getEnv(key, defaultValue string) string {
	raw := os.Getenv("VIRTUALPAPER_" + strings.ToUpper(key))
	if raw == "" {
		return defaultValue
	}
	return raw
}
