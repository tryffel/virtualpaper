package config

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"math/rand"
)

// Application config that is loaded upon starting program
var C *Config

type Config struct {
	Api         Api
	Database    Database
	Preferences Preferences

	BackgroundTasks BackgroundTasks
}

type Api struct {
	Host string
	Port int
	Key  string
}

type Database struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
}

type Preferences struct {
}

type BackgroundTasks struct {
	UserCheckedIn     bool
	USerCheckedInTime string
}

// ConfigFromViper initializes Config.C, reads all config values from viper and stores them to Config.C.
func ConfigFromViper() error {

	c := &Config{
		Api: Api{
			Host: viper.GetString("api.host"),
			Port: viper.GetInt("api.port"),
			Key:  viper.GetString("api.secret_key"),
		},

		Database: Database{
			Host:     viper.GetString("database.host"),
			Port:     viper.GetInt("database.Port"),
			Username: viper.GetString("database.username"),
			Password: viper.GetString("database.password"),
			Database: viper.GetString("database.database"),
		},
		Preferences: Preferences{},
		BackgroundTasks: BackgroundTasks{
			UserCheckedIn:     viper.GetBool("backgroundtasks.user_checked_in"),
			USerCheckedInTime: viper.GetString("backgroundtasks.user_checked_in_time"),
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
