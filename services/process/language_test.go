package process

import (
	"context"
	"testing"
	log "tryffel.net/go/virtualpaper/util/logger"
)

var samples = [][]string{
	{"en", "languages are awesome"},
	{"en", "a much longer Variant OF TEXTs are represented here"},
	{"de", "mixed auf Zeitung"},
}

func TestGetLanguage(t *testing.T) {
	initLanguageDetector()
	ctx := log.ContextWithTaskId(context.Background(), "test-task")
	for i, sample := range samples {
		lang, _ := detectLanguage(ctx, sample[1])
		if sample[0] != lang {
			t.Errorf("case %d: did not detect language, got '%s', want: '%s'", i+1, lang, sample[0])
		}
	}
}
