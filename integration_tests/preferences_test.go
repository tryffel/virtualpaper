package integrationtest

import (
	"context"
	"fmt"
	"github.com/meilisearch/meilisearch-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
	"tryffel.net/go/virtualpaper/api"
	"tryffel.net/go/virtualpaper/models"
)

type PreferencesTest struct {
	ApiTestSuite
	users map[string]models.User
}

func (suite *PreferencesTest) SetupTest() {
	suite.Init()
	clearTestUsersTables(suite.T(), suite.db)

	users, err := suite.db.UserStore.GetUsers()
	assert.Nil(suite.T(), err)

	userIds := make([]int, len(*users))
	suite.users = map[string]models.User{}
	for i, v := range *users {
		userIds[i] = v.Id
		suite.users[v.Name] = v
	}
}

func (suite *PreferencesTest) TearDownSuite() {
	clearTestUsersTables(suite.T(), suite.db)
	suite.ApiTestSuite.TearDownSuite()
}

func TestPreferences(t *testing.T) {
	suite.Run(t, new(PreferencesTest))
}

func getMeilisearchPreferences(t *testing.T, userId int) (stopWords *[]string, synonyms *map[string][]string) {
	client := meilisearch.NewClient(meilisearch.ClientConfig{
		Host:    meiliHost,
		APIKey:  meilisearchKey,
		Timeout: 2 * time.Second,
	})
	var err error

	stopWords, err = client.Index(fmt.Sprintf("virtualpaper-%d", userId)).GetStopWords()
	assert.Nil(t, err)
	synonyms, err = client.Index(fmt.Sprintf("virtualpaper-%d", userId)).GetSynonyms()
	assert.Nil(t, err)
	return
}

func clearMeilisearchPreferences(t *testing.T, userIds []int) []int64 {
	client := meilisearch.NewClient(meilisearch.ClientConfig{
		Host:    meiliHost,
		APIKey:  meilisearchKey,
		Timeout: 2 * time.Second,
	})

	taskIds := make([]int64, 0, 2*len(userIds))

	for _, v := range userIds {
		task0, err := client.Index(fmt.Sprintf("virtualpaper-%d", v)).ResetStopWords()
		assert.Nil(t, err)
		task1, err := client.Index(fmt.Sprintf("virtualpaper-%d", v)).ResetSynonyms()
		assert.Nil(t, err)
		taskIds = append(taskIds, task0.TaskUID, task1.TaskUID)
	}

	pollUntilMeilisearchTaskReady(t, taskIds)
	return taskIds
}

func pollUntilMeilisearchTaskReady(t *testing.T, taskIds []int64) {
	client := meilisearch.NewClient(meilisearch.ClientConfig{
		Host:    meiliHost,
		APIKey:  meilisearchKey,
		Timeout: 2 * time.Second,
	})

	lastTaskid := taskIds[(len(taskIds) - 1)]
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	t.Log("Start polling meilisearch")

	_, err := client.WaitForTask(lastTaskid, meilisearch.WaitParams{
		Context:  ctx,
		Interval: time.Millisecond * 10,
	})
	t.Log("Poll completed")
	assert.Nil(t, err)
}

func updateUserPreferences(t *testing.T, client *httpClient, pref *api.ReqUserPreferences, wantHttpStatus int) {
	req := client.Put("/api/v1/preferences/user").Json(t, pref).Expect(t)
	req.e.Status(wantHttpStatus).Done()
}

func getUserPreferences(t *testing.T, client *httpClient, wantHttpStatus int) *api.UserPreferences {
	req := client.Get("/api/v1/preferences/user").Expect(t)
	dto := &api.UserPreferences{}
	if wantHttpStatus == 200 {
		req.Json(t, dto).e.Status(wantHttpStatus).Done()
		return dto
	} else {
		req.e.Status(wantHttpStatus).Done()
		return nil
	}
}
