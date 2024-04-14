package integrationtest

import (
	"database/sql"
	"strings"
	"testing"
	"time"
	"tryffel.net/go/virtualpaper/models"
	"tryffel.net/go/virtualpaper/storage"
)

var defaultDate = time.Date(2023, 1, 21, 13, 5, 0, 0, time.UTC)

func newDocument(id string, name string, content string, userId int) *models.Document {
	return &models.Document{
		Timestamp: models.Timestamp{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Id:          id,
		UserId:      userId,
		Name:        name,
		Description: "description - " + name,
		Content:     content,
		Filename:    "",
		Hash:        strings.Trim(id, "-"),
		Mimetype:    "application/pdf",
		Size:        1024,
		Date:        defaultDate,
		Metadata:    nil,
		Tags:        nil,
		DeletedAt:   sql.NullTime{},
	}
}

var testDocumentX86 = newDocument("c6efd246-dc38-4a58-ad53-8bdea88f88a7", "x86 architecture",
	`x86 is the name of a instruction set architecture (ISA),
which is a set of instructions that a computer's processor can execute. 
The x86 ISA is a type of CISC (Complex Instruction Set Computing) architecture,
which means that it includes a large number of instructions that can perform a wide variety of operations.

05/06/2023
The architecture is also known for its backwards compatibility feature,
which means that newer processors are able to run software written for older processors without modification,
this feature enables the architecture to maintain a large software ecosystem.
`, 1)

var testDocumentX86Intel = newDocument("d8579eea-47d9-4d5c-9892-2f6a4301b1a2", "x86 intel",
	`The x86 architecture was first developed by Intel in 1978 for use in its microprocessors.
Since then, it has become one of the most widely used ISAs in the world, 
and is found in many personal computers, servers, and other devices.`, 1)

var testDocumentJupiterMoons = newDocument("23eb40f2-bb23-41aa-8c8e-290d9028bce6", "Jupiter's moons",
	`Jupiter has 79 known moons. 
The four largest moons, known as the Galilean moons, were discovered by Galileo Galilei in 1610 and are named after him: Io, Europa, Ganymede, and Callisto. 
These four moons are some of the largest objects in the solar system outside of the Sun and the eight planets, 
and they are also some of the most geologically active bodies in the solar system.`, 1)

var testDocumentYear1962 = newDocument("4fcb479e-085e-4c28-ba5f-0794762c67fd", "Year 1962",
	`1962 was a significant year in history, several important events occurred:

The first communications satellite, Telstar, was launched into orbit, 
allowing live television broadcasts to be transmitted across the Atlantic for the first time.`, 1)

var testDocumentMetamorphosis = newDocument("c509084c-6fcd-4bd2-b2dd-f7b11fdaaad9", "metamorphosis",
	`One example of cell metamorphosis occurs during the development of animals, such as insects. 
Insects go through a process called metamorphosis during which they change from an immature form, such as a caterpillar, to an adult form, such as a butterfly. 
This process is known as holometabolous metamorphosis, and it involves a radical change in the insect's body shape and physiology. 
During this process, cells in the caterpillar's body undergo metamorphosis, transforming into the cells that make up the adult butterfly.`, 1)

var testDocumentTransistorCount = newDocument("ed9751b3-d0b5-4485-860c-772c7d1e79cb", "transistor count",
	`Transistor count is a rough measure of the complexity and capabilities of a computer's processor.
The more transistors a processor has, the more calculations it can perform per second and the more powerful it is.`, 1)

var testDocumentTransistorCountAdminUser = newDocument("21eb96fd-4475-40a4-992f-6150217ffe9e ", "transistor count",
	`Transistor count is a rough measure of the complexity and capabilities of a computer's processor.
The more transistors a processor has, the more calculations it can perform per second and the more powerful it is.`, 2)

var testDocuments = []*models.Document{
	testDocumentX86,
	testDocumentX86Intel,
	testDocumentJupiterMoons,
	testDocumentMetamorphosis,
	testDocumentYear1962,
	testDocumentTransistorCount,
	testDocumentTransistorCountAdminUser,
}

var testDocumentIds = []string{
	testDocumentTransistorCountAdminUser.Id,
	testDocumentTransistorCount.Id,
	testDocumentYear1962.Id,
	testDocumentMetamorphosis.Id,
	testDocumentJupiterMoons.Id,
	testDocumentX86Intel.Id,
	testDocumentX86.Id,
}

var testDocumentIdsUser = []string{
	testDocumentTransistorCount.Id,
	testDocumentYear1962.Id,
	testDocumentMetamorphosis.Id,
	testDocumentJupiterMoons.Id,
	testDocumentX86Intel.Id,
	testDocumentX86.Id,
}

var testDocumentIdsAdmin = []string{
	testDocumentTransistorCountAdminUser.Id,
}

func insertTestDocuments(t *testing.T, db *storage.Database) error {
	for _, v := range testDocuments {
		time.Sleep(time.Millisecond)
		err := db.DocumentStore.Create(db, v)
		if err != nil {
			t.Errorf("insert test document %s: %v", v.Id, err)
			t.Fail()
			return err
		}
	}
	return nil
}

var testMetadataKeys = &[]models.MetadataKey{
	{
		Key:     "Author",
		Comment: "author",
	},
	{
		Key:     "Subject",
		Comment: "subject",
	},
	{
		Key:     "Language",
		Comment: "language",
	},
	{
		Key:     "Publisher",
		Comment: "publisher",
	},
	{
		Key:     "Location",
		Comment: "location",
	},
	{
		Key:     "Archived",
		Comment: "archived",
	},
	{
		Key:     "Status",
		Comment: "status",
	},
}

var testMetadataValues = &map[string][]models.MetadataValue{
	"Author": {},
}
