package configuration

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/rancoud/blueprintue-discord/welcome"
)

// Support only field's type string, int, bool, []string

type Configuration struct {
	Discord struct {
		Name  string `json:"name" env:"BUE_BOT_DISCORD_NAME"`
		Token string `json:"token" env:"BUE_BOT_DISCORD_TOKEN"`
	} `json:"discord"`
	Twitch struct {
		ClientID string `json:"clientID" env:"BUE_BOT_TWITCH_CLIENT_ID"`
		UserID   string `json:"userID" env:"BUE_BOT_TWITCH_USER_ID"`
	} `json:"twitch"`
	Log struct {
		Filename string `json:"filename" env:"BUE_BOT_LOG_FILENAME"`
		Level    string `json:"level" env:"BUE_BOT_LOG_LEVEL"`
	}
	Modules struct {
		WelcomeConfiguration welcome.Configuration `json:"welcome"`
	} `json:"modules"`
}

func ReadConfiguration(filename string) (*Configuration, error) {
	filedata, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	config := Configuration{}
	err = json.Unmarshal(filedata, &config)
	if err != nil {
		return nil, err
	}

	eraseConfigurationValuesWithEnv(&config)

	return &config, nil
}

func eraseConfigurationValuesWithEnv(obj interface{}) interface{} {
	var val reflect.Value
	if reflect.TypeOf(obj).Kind() == reflect.Ptr {
		val = reflect.ValueOf(obj).Elem()
	} else {
		val = reflect.ValueOf(obj)
	}

	for idxNumField := 0; idxNumField < val.NumField(); idxNumField++ {
		if val.Field(idxNumField).Kind() == reflect.Struct {
			eraseConfigurationValuesWithEnv(val.Field(idxNumField).Addr().Interface())
			continue
		}

		envKey := reflect.TypeOf(obj).Elem().Field(idxNumField).Tag.Get("env")
		envValue, ok := os.LookupEnv(envKey)
		if !ok {
			continue
		}

		switch val.Field(idxNumField).Kind() {
		case reflect.String:
			val.Field(idxNumField).SetString(envValue)
		case reflect.Int:
			intEnvValue, err := strconv.Atoi(envValue)
			if err == nil {
				val.Field(idxNumField).SetInt(int64(intEnvValue))
			}
		case reflect.Bool:
			boolEnvValue, err := strconv.ParseBool(envValue)
			if err == nil {
				val.Field(idxNumField).SetBool(boolEnvValue)
			}
		case reflect.Slice:
			splitter := ":"
			if runtime.GOOS == "windows" {
				splitter = ";"
			}
			stringEnvValues := strings.Split(envValue, splitter)
			val.Field(idxNumField).Set(reflect.MakeSlice(val.Field(idxNumField).Type(), len(stringEnvValues), len(stringEnvValues)))
			for idxSlice := 0; idxSlice < len(stringEnvValues); idxSlice++ {
				val.Field(idxNumField).Index(idxSlice).SetString(stringEnvValues[idxSlice])
			}
		}
	}

	return obj
}
