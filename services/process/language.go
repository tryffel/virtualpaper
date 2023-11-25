package process

import (
	"context"
	"fmt"
	"github.com/pemistahl/lingua-go"
	"github.com/sirupsen/logrus"
	"time"
	"tryffel.net/go/virtualpaper/errors"
	"tryffel.net/go/virtualpaper/models"
	"tryffel.net/go/virtualpaper/storage"
	log "tryffel.net/go/virtualpaper/util/logger"
	"unicode"
)

var languageDetector lingua.LanguageDetector

func (fp *fileProcessor) detectLanguage(ctx context.Context) error {
	process := &models.ProcessItem{
		DocumentId: fp.document.Id,
		Action:     models.ProcessDetectLanguage,
		CreatedAt:  time.Now(),
	}
	job, err := fp.db.JobStore.StartProcessItem(process, "detect language")
	// hotfix for failure when job item does not exist anymore.
	if err != nil {
		logrus.Warningf("persist job record: %v", err)
		// use empty job to not panic the rest of the function
		job = &models.Job{}
	} else {
		defer fp.completeProcessingStep(process, job)
	}

	lang, err := detectLanguage(ctx, fp.document.Content)
	if err != nil {
		job.Status = models.JobFailure
		return err
	}
	if lang != "" {
		fp.document.Lang = models.Lang(lang)
	}
	job.Status = models.JobFinished
	logrus.Debugf("Detected language: %s", lang)

	err = fp.db.DocumentStore.Update(storage.UserIdInternal, fp.document)
	if err != nil {
		logrus.Errorf("update document (%s) after rules: %v", fp.document.Id, err)
	}
	return nil
}

func detectLanguage(ctx context.Context, text string) (string, error) {
	maxLen := 500
	truncated := text
	if len(text) > maxLen {
		truncated = text[:500]
	}

	cleanedUp := make([]rune, 0, len(truncated))
	for _, v := range truncated {
		if unicode.IsLetter(v) {
			cleanedUp = append(cleanedUp, v)
		} else if unicode.IsSymbol(v) || unicode.IsSpace(v) {
			cleanedUp = append(cleanedUp, ' ')
		}
	}

	log.Context(ctx).Info("detect lang")
	sample := string(cleanedUp)
	lang, found := languageDetector.DetectLanguageOf(sample)
	if !found {
		return "", nil
	}
	langCode := supportedLinguaLanguages[lang]
	if langCode == "" {
		err := errors.ErrInternalError
		err.ErrMsg = fmt.Sprintf("detected unsupported language: %v", lang)
		return "", err
	}
	return langCode, nil
}

func initLanguageDetector() {
	languageDetector = lingua.NewLanguageDetectorBuilder().
		FromAllLanguages().
		WithMinimumRelativeDistance(0.15).
		Build()
}

var SupportedLanguages = map[string]string{
	"af": "Afrikaans",
	"sq": "Albanian",
	"ar": "Arabic",
	"hy": "Armenian",
	"az": "Azerbaijani",
	"eu": "Basque",
	"be": "Belarusian",
	"bn": "Bengali",
	"nb": "Bokmal",
	"bs": "Bosnian",
	"bg": "Bulgarian",
	"ca": "Catalan",
	"zh": "Chinese",
	"hr": "Croatian",
	"cs": "Czech",
	"da": "Danish",
	"nl": "Dutch",
	"en": "English",
	"eo": "Esperanto",
	"et": "Estonian",
	"fi": "Finnish",
	"fr": "French",
	"lg": "Ganda",
	"ka": "Georgian",
	"de": "German",
	"el": "Greek",
	"gu": "Gujarati",
	"he": "Hebrew",
	"hi": "Hindi",
	"hu": "Hungarian",
	"is": "Icelandic",
	"id": "Indonesian",
	"ga": "Irish",
	"it": "Italian",
	"ja": "Japanese",
	"kk": "Kazakh",
	"ko": "Korean",
	"la": "Latin",
	"lv": "Latvian",
	"lt": "Lithuanian",
	"mk": "Macedonian",
	"ms": "Malay",
	"mi": "Maori",
	"mr": "Marathi",
	"mn": "Mongolian",
	"nn": "Nynorsk",
	"fa": "Persian",
	"pl": "Polish",
	"pt": "Portuguese",
	"pa": "Punjabi",
	"rm": "Romanian",
	"ru": "Russian",
	"sr": "Serbian",
	"sn": "Shona",
	"sk": "Slovak",
	"sl": "Slovene",
	"so": "Somali",
	"st": "Sotho",
	"es": "Spanish",
	"sw": "Swahili",
	"sv": "Swedish",
	"tl": "Tagalog",
	"ta": "Tamil",
	"te": "Telugu",
	"th": "Thai",
	"ts": "Tsonga",
	"tn": "Tswana",
	"tr": "Turkish",
	"uk": "Ukrainian",
	"ur": "Urdu",
	"vi": "Vietnamese",
	"cy": "Welsh",
	"xh": "Xhosa",
	"yo": "Yoruba",
	"zu": "Zulu",
}

var supportedLinguaLanguages = map[lingua.Language]string{
	lingua.Afrikaans:   "af",
	lingua.Albanian:    "sq",
	lingua.Arabic:      "ar",
	lingua.Armenian:    "hy",
	lingua.Azerbaijani: "az",
	lingua.Basque:      "eu",
	lingua.Belarusian:  "be",
	lingua.Bengali:     "bn",
	lingua.Bokmal:      "nb",
	lingua.Bosnian:     "bs",
	lingua.Bulgarian:   "bg",
	lingua.Catalan:     "ca",
	lingua.Chinese:     "zh",
	lingua.Croatian:    "hr",
	lingua.Czech:       "cs",
	lingua.Danish:      "da",
	lingua.Dutch:       "nl",
	lingua.English:     "en",
	lingua.Esperanto:   "eo",
	lingua.Estonian:    "et",
	lingua.Finnish:     "fi",
	lingua.French:      "fr",
	lingua.Ganda:       "lg",
	lingua.Georgian:    "ka",
	lingua.German:      "de",
	lingua.Greek:       "el",
	lingua.Gujarati:    "gu",
	lingua.Hebrew:      "he",
	lingua.Hindi:       "hi",
	lingua.Hungarian:   "hu",
	lingua.Icelandic:   "is",
	lingua.Indonesian:  "id",
	lingua.Irish:       "ga",
	lingua.Italian:     "it",
	lingua.Japanese:    "ja",
	lingua.Kazakh:      "kk",
	lingua.Korean:      "ko",
	lingua.Latin:       "la",
	lingua.Latvian:     "lv",
	lingua.Lithuanian:  "lt",
	lingua.Macedonian:  "mk",
	lingua.Malay:       "ms",
	lingua.Maori:       "mi",
	lingua.Marathi:     "mr",
	lingua.Mongolian:   "mn",
	lingua.Nynorsk:     "nn",
	lingua.Persian:     "fa",
	lingua.Polish:      "pl",
	lingua.Portuguese:  "pt",
	lingua.Punjabi:     "pa",
	lingua.Romanian:    "rm",
	lingua.Russian:     "ru",
	lingua.Serbian:     "sr",
	lingua.Shona:       "sn",
	lingua.Slovak:      "sk",
	lingua.Slovene:     "sl",
	lingua.Somali:      "so",
	lingua.Sotho:       "st",
	lingua.Spanish:     "es",
	lingua.Swahili:     "sw",
	lingua.Swedish:     "sv",
	lingua.Tagalog:     "tl",
	lingua.Tamil:       "ta",
	lingua.Telugu:      "te",
	lingua.Thai:        "th",
	lingua.Tsonga:      "ts",
	lingua.Tswana:      "tn",
	lingua.Turkish:     "tr",
	lingua.Ukrainian:   "uk",
	lingua.Urdu:        "ur",
	lingua.Vietnamese:  "vi",
	lingua.Welsh:       "cy",
	lingua.Xhosa:       "xh",
	lingua.Yoruba:      "yo",
	lingua.Zulu:        "zu",
}
