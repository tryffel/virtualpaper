package process

import "testing"

var samples = [][]string{
	{"en", "languages are awesome"},
	{"en", "a much longer Variant OF TEXTs are represented here"},
	{"de", "mixed auf Zeitung"},
}

func TestGetLanguage(t *testing.T) {
	initLanguageDetector()
	for i, sample := range samples {
		lang, _ := detectLanguage(sample[1])
		if sample[0] != lang {
			t.Errorf("case %d: did not detect language, got '%d' '%s', want: '%s'", i+1, lang, sample[0])
		}
	}
}
