package integrationtest

import (
	"fmt"
	"github.com/Masterminds/squirrel"
	"testing"
	"tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/services/search"
	"tryffel.net/go/virtualpaper/storage"
)

func GetDb() *storage.Database {
	// docker-compose.yml
	conf := config.Database{
		Host:     dbHost,
		Port:     5432,
		Username: "virtualpaper",
		Password: "virtualpaper",
		Database: "virtualpaper",
		NoSSL:    true,
	}

	db, err := storage.NewDatabase(conf)
	if err != nil {
		panic(fmt.Errorf("init db: %v", err))
	}
	return db
}

func closeDb(db *storage.Database, t *testing.T) {
	err := db.Close()
	if err != nil {
		t.Errorf("close db connection: %v", err)
	}
}

var dbMetadataTables = []string{
	"documents",
	"metadata_values",
	"metadata_keys",
}

var dbProcessingRuleTables = []string{
	"rule_actions",
	"rule_conditions",
	"rules",
}

var dbDocumentTables = []string{
	"document_view_history",
	"document_history",
	"linked_documents",
	"jobs",
	"documents",
}

var dbPasswordResetTables = []string{
	"password_reset_tokens",
}

func clearDbMetadataTables(t *testing.T, db *storage.Database) {
	for _, v := range dbMetadataTables {
		db.Engine().MustExec(fmt.Sprintf("DELETE FROM %s WHERE 1=1", v))
	}
}

func clearDbProcessingRuleTables(t *testing.T, db *storage.Database) {
	for _, v := range dbProcessingRuleTables {
		db.Engine().MustExec(fmt.Sprintf("DELETE FROM %s WHERE 1=1", v))
	}
}

func clearDbDocumentTables(t *testing.T, db *storage.Database) {
	for _, v := range dbDocumentTables {
		db.Engine().MustExec(fmt.Sprintf("DELETE FROM %s WHERE 1=1", v))
	}
}

func clearPasswordResetTables(t *testing.T, db *storage.Database) {
	for _, v := range dbPasswordResetTables {
		db.Engine().MustExec(fmt.Sprintf("DELETE FROM %s WHERE 1=1", v))
	}
}

func clearTestUsersTables(t *testing.T, db *storage.Database) {
	users, err := db.UserStore.GetUsers()
	if err != nil {
		t.Errorf("get users: %v", err)
		return
	}

	ids := make([]int, 0, 10)
	for _, user := range *users {
		if user.Name != "user" && user.Name != "admin" && user.Name != "tester" {
			ids = append(ids, user.Id)
		}
	}
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query := builder.Delete("user_preferences").Where(squirrel.Eq{"user_id": ids})
	sql, args, err := query.ToSql()
	if err != nil {
		t.Errorf("build delete query: %v", err)
	}
	db.Engine().MustExec(sql, args...)

	query = builder.Delete("users").Where(squirrel.Eq{"id": ids})
	sql, args, err = query.ToSql()
	if err != nil {
		t.Errorf("build delete query: %v", err)
	}
	db.Engine().MustExec(sql, args...)
}

func clearMeiliIndices(t *testing.T) {
	db := GetDb()
	defer closeDb(db, t)

	conf := &config.Meilisearch{
		Url:    meiliHost,
		Index:  "virtualpaper",
		ApiKey: "",
	}

	client, err := search.NewEngine(db, conf)
	if err != nil {
		t.Error("connect to Meilisearch", err)
	}

	users, err := db.UserStore.GetUsers()
	if err != nil {
		t.Error("get users from db", err)
	}

	for _, v := range *users {
		err := client.DeleteDocuments(v.Id)
		if err != nil {
			t.Logf("delete search index for user %d: %v", v.Id, err)
		}
	}
}
