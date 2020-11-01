package config

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"math/rand"
	"os"
	"path"
	"runtime"
	"strings"
)

// Application config that is loaded upon starting program
var C *Config

// Config contains application config
type Config struct {
	Api         Api
	Database    Database
	Processing  Processing
	Meilisearch Meilisearch
}

// Api contains http server config
type Api struct {
	Host string
	Port int
	Key  string

	PublicUrl string
	CorsHosts []string
}

func (a *Api) CorsHostList() string {
	return strings.Join(a.CorsHosts, ",")
}

// Database
type Database struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
}

// Processing contains document-processing settings
type Processing struct {
	InputDir     string
	TmpDir       string
	DataDir      string
	MaxWorkers   int
	OcrLanguages []string
	PdfToTextBin string

	// application directories. Stored by default in ./media/{previews, documents}.
	PreviewsDir  string
	DocumentsDir string
}

type Meilisearch struct {
	Url    string
	Index  string
	ApiKey string
}

// ConfigFromViper initializes Config.C, reads all config values from viper and stores them to Config.C.
func ConfigFromViper() error {

	c := &Config{
		Api: Api{
			Host:      viper.GetString("api.host"),
			Port:      viper.GetInt("api.port"),
			Key:       viper.GetString("api.secret_key"),
			PublicUrl: viper.GetString("api.public_url"),
			CorsHosts: viper.GetStringSlice("api.cors_hosts"),
		},

		Database: Database{
			Host:     viper.GetString("database.host"),
			Port:     viper.GetInt("database.Port"),
			Username: viper.GetString("database.username"),
			Password: viper.GetString("database.password"),
			Database: viper.GetString("database.database"),
		},
		Processing: Processing{
			InputDir:     viper.GetString("processing.input_dir"),
			TmpDir:       viper.GetString("processing.tmp_dir"),
			DataDir:      viper.GetString("processing.data_dir"),
			MaxWorkers:   viper.GetInt("processing.max_workers"),
			OcrLanguages: viper.GetStringSlice("processing.ocr_languages"),
			PdfToTextBin: viper.GetString("processing.pdftotext_bin"),
		},
		Meilisearch: Meilisearch{
			Url:    viper.GetString("meilisearch.url"),
			Index:  viper.GetString("meilisearch.index"),
			ApiKey: viper.GetString("meilisearch.apikey"),
		},
	}

	var err error

	C = c
	return err
}

// InitConfig sets sane default values and creates necessary keys. This can be called only after initializing Config.C.
func InitConfig() error {
	changed := false

	if C.Api.Key == "" {
		logrus.Info("create api key of 50 characters")
		C.Api.Key = RandomString(50)
		viper.Set("api.secret_key", C.Api.Key)
		changed = true
	}

	var inputChanged, tmpChanged, dataChanged, indexChanged bool

	C.Processing.InputDir, inputChanged = setVar(C.Processing.InputDir, "input")

	defaultTmpDir := os.TempDir()
	defaultTmpDir = path.Join(defaultTmpDir, "virtualpaper")
	C.Processing.TmpDir, inputChanged = setVar(C.Processing.TmpDir, defaultTmpDir)
	C.Processing.DataDir, dataChanged = setVar(C.Processing.DataDir, "data")
	C.Meilisearch.Index, indexChanged = setVar(C.Meilisearch.Index, "virtualpaper")
	if len(C.Processing.OcrLanguages) == 0 {
		C.Processing.OcrLanguages = []string{"eng"}
		viper.Set("processing.ocr_languages", C.Processing.OcrLanguages)
		changed = true
	}

	if !path.IsAbs(C.Processing.DataDir) {
		curDir, err := os.Getwd()
		if err != nil {
			logrus.Error("cannot determine current directory. Please set processing.output_dir as absolute directory")
		} else {
			C.Processing.DataDir = path.Join(curDir, C.Processing.DataDir)
		}
	}

	C.Processing.DocumentsDir = path.Join(C.Processing.DataDir, "documents")
	C.Processing.PreviewsDir = path.Join(C.Processing.DataDir, "previews")

	viper.Set("processing.tmp_dir", C.Processing.TmpDir)
	viper.Set("processing.data_dir", C.Processing.DataDir)
	viper.Set("processing.input_dir", C.Processing.InputDir)

	if C.Processing.MaxWorkers == 0 {
		C.Processing.MaxWorkers = runtime.NumCPU()
	}

	err := os.MkdirAll(C.Processing.DataDir, os.ModePerm)
	if err != nil {
		logrus.Errorf("create data directory: %v", err)
	}

	err = os.MkdirAll(C.Processing.TmpDir, os.ModePerm)
	if err != nil {
		logrus.Errorf("create tmp directory: %v", err)
	}

	err = os.Mkdir(C.Processing.PreviewsDir, os.ModePerm)
	if err != nil {
		if !errors.Is(err, os.ErrExist) {
			logrus.Errorf("create previews directory: %v", err)
		}
	}

	err = os.Mkdir(C.Processing.DocumentsDir, 777)
	if err != nil {
		if !errors.Is(err, os.ErrExist) {
			logrus.Errorf("create documents directory: %v", err)
		}
	}

	viper.Set("processing.max_workers", C.Processing.MaxWorkers)
	changed = changed || inputChanged || tmpChanged || dataChanged || indexChanged
	if changed {
		err := viper.WriteConfig()
		if err != nil {
			return fmt.Errorf("save config file: %v", err)
		}
	}
	return nil
}

// RandomString creates new random string of given length in characters.
// Modified from https://socketloop.com/tutorials/golang-derive-cryptographic-key-from-passwords-with-argon2
func RandomString(size int) string {
	dict := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, size)
	rand.Read(bytes)
	for k, v := range bytes {
		bytes[k] = dict[v%byte(len(dict))]
	}
	return string(bytes)
}

// setVar returns currentVal and false if currentVal is not "", else return newVal and true
func setVar(currentVal, newVal string) (string, bool) {
	if currentVal != "" {
		return currentVal, false
	}
	return newVal, true
}
