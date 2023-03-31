package config

import (
	crypto "crypto/rand"
	"errors"
	"fmt"
	math "math/rand"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Application config that is loaded upon starting program
var C *Config

// Config contains application config
type Config struct {
	Api         Api
	Database    Database
	Processing  Processing
	Meilisearch Meilisearch
	Mail        Mail
	Logging     Logging
	CronJobs    CronJobs
}

// Api contains http server config
type Api struct {
	Host           string
	Port           int
	Key            string
	TokenExpireSec int
	TokenExpire    time.Duration

	PublicUrl string
	CorsHosts []string

	StaticContentPath     string
	AuthRatelimitDisabled bool
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
	NoSSL    bool
}

// Processing contains document-processing settings
type Processing struct {
	Disabled     bool
	InputDir     string
	TmpDir       string
	DataDir      string
	MaxWorkers   int
	OcrLanguages []string
	PdfToTextBin string
	PandocBin    string
	ImagickBin   string
	TesseractBin string

	// application directories. Stored by default in ./media/{previews, documents}.
	PreviewsDir  string
	DocumentsDir string
}

// Meilisearch contains search-engine configuration
type Meilisearch struct {
	Url    string
	Index  string
	ApiKey string
}

// Mail contains configuration for sending mails.
type Mail struct {
	// Is mailing enabled
	Enabled bool

	// Smpt host
	Host string
	Port int

	Username string
	Password string

	From string
	// Recipient to send errors
	ErrorRecipient string
}

// Logging configuration
type Logging struct {
	Loglevel      string
	LogDirectory  string
	LogHttpStdout bool
	HttpLogFile   string
	LogFile       string
	LogStdout     bool

	httpLog *os.File
	log     *os.File

	HttpLog *logrus.Logger
}

type CronJobs struct {
	Disabled                  bool
	DocumentsTrashbinDuration time.Duration
}

// ConfigFromViper initializes Config.C, reads all config values from viper and stores them to Config.C.
func ConfigFromViper() error {

	c := &Config{
		Api: Api{
			Host:                  viper.GetString("api.host"),
			Port:                  viper.GetInt("api.port"),
			Key:                   viper.GetString("api.secret_key"),
			PublicUrl:             viper.GetString("api.public_url"),
			CorsHosts:             viper.GetStringSlice("api.cors_hosts"),
			StaticContentPath:     viper.GetString("api.static_content_path"),
			TokenExpireSec:        viper.GetInt("api.token_expire_sec"),
			AuthRatelimitDisabled: viper.GetBool("api.disable_auth_ratelimit"),
		},

		Database: Database{
			Host:     viper.GetString("database.host"),
			Port:     viper.GetInt("database.port"),
			Username: viper.GetString("database.username"),
			Password: viper.GetString("database.password"),
			Database: viper.GetString("database.database"),
			NoSSL:    viper.GetBool("database.no_ssl"),
		},
		Processing: Processing{
			Disabled:     viper.GetBool("processing.disabled"),
			InputDir:     viper.GetString("processing.input_dir"),
			TmpDir:       viper.GetString("processing.tmp_dir"),
			DataDir:      viper.GetString("processing.data_dir"),
			MaxWorkers:   viper.GetInt("processing.max_workers"),
			OcrLanguages: viper.GetStringSlice("processing.ocr_languages"),
			PdfToTextBin: viper.GetString("processing.pdftotext_bin"),
			PandocBin:    viper.GetString("processing.pandoc_bin"),
			ImagickBin:   viper.GetString("processing.imagick_bin"),
			TesseractBin: viper.GetString("processing.tesseract_bin"),
		},
		Meilisearch: Meilisearch{
			Url:    viper.GetString("meilisearch.url"),
			Index:  viper.GetString("meilisearch.index"),
			ApiKey: viper.GetString("meilisearch.apikey"),
		},
		Mail: Mail{
			Enabled:        false,
			Host:           viper.GetString("mail.host"),
			Port:           viper.GetInt("mail.port"),
			Username:       viper.GetString("mail.username"),
			Password:       viper.GetString("mail.password"),
			From:           viper.GetString("mail.from"),
			ErrorRecipient: viper.GetString("mail.error_recipient"),
		},
		Logging: Logging{
			Loglevel:      viper.GetString("logging.log_level"),
			LogDirectory:  viper.GetString("logging.directory"),
			LogHttpStdout: viper.GetBool("logging.log_http_stdout"),
			HttpLogFile:   viper.GetString("logging.http_log_file"),
			LogFile:       viper.GetString("logging.log_file"),
			LogStdout:     viper.GetBool("logging.log_stdout"),
		},
		CronJobs: CronJobs{
			Disabled:                  viper.GetBool("cronjobs.disabled"),
			DocumentsTrashbinDuration: viper.GetDuration("cronjobs.documents_trashbin_cleanup_duration"),
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
		var err error
		C.Api.Key, err = RandomStringCrypt(200)
		if err != nil {
			return fmt.Errorf("generate api key: %v", err)
		}
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

	if C.Api.TokenExpireSec != 0 {
		C.Api.TokenExpire = time.Second * time.Duration(C.Api.TokenExpireSec)
	}

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

	viper.Set("logging.log_level", C.Logging.Loglevel)
	viper.Set("logging.directory", C.Logging.LogDirectory)
	viper.Set("logging.log_http_stdout", C.Logging.LogHttpStdout)
	viper.Set("logging.http_log_file", C.Logging.HttpLogFile)
	viper.Set("logging.log_file", C.Logging.LogFile)

	if C.Processing.MaxWorkers == 0 {
		// use only half of available cpus
		C.Processing.MaxWorkers = runtime.NumCPU() / 2
		if C.Processing.MaxWorkers == 0 {
			C.Processing.MaxWorkers = 1
		}
	}

	if C.Mail.Host != "" {
		C.Mail.Enabled = true
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

	err = os.Mkdir(C.Processing.DocumentsDir, os.ModePerm)
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

// RandomStringCrypt creates new random string of given length in characters (cryptographic).
// Modified from https://socketloop.com/tutorials/golang-derive-cryptographic-key-from-passwords-with-argon2
func RandomStringCrypt(size int) (string, error) {
	dict := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, size)
	_, err := crypto.Read(bytes)
	for k, v := range bytes {
		bytes[k] = dict[v%byte(len(dict))]
	}
	return string(bytes), err
}

// RandomString creates new random string of given length in characters (not cryptographic).
// Modified from https://socketloop.com/tutorials/golang-derive-cryptographic-key-from-passwords-with-argon2
func RandomString(size int) (string, error) {
	dict := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, size)
	_, err := math.Read(bytes)
	for k, v := range bytes {
		bytes[k] = dict[v%byte(len(dict))]
	}
	return string(bytes), err
}

// setVar returns currentVal and false if currentVal is not "", else return newVal and true
func setVar(currentVal, newVal string) (string, bool) {
	if currentVal != "" {
		return currentVal, false
	}
	return newVal, true
}
