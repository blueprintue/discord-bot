package configuration

import (
	"encoding/json"
	"github.com/davecgh/go-spew/spew"
	"io/ioutil"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
)

// Support only field's type string, int, bool, []string

type Configuration struct {
	Discord struct {
		Channel string   `json:"channel" env:"BUE_BOT_DISCORD_CHANNELS"`
		Roles   string   `json:"roles"   env:"BUE_BOT_DISCORD_ROLES"`
		Token   string   `json:"token"   env:"BUE_BOT_DISCORD_TOKEN"`
	} `json:"discord"`
	Twitch struct {
		ClientID string `json:"clientID" env:"BUE_BOT_TWITCH_CLIENT_ID"`
		UserID string `json:"userID" env:"BUE_BOT_TWITCH_USER_ID"`
	} `json:"twitch"`
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
			break
		case reflect.Int:
			intEnvValue, err := strconv.Atoi(envValue)
			if err == nil {
				val.Field(idxNumField).SetInt(int64(intEnvValue))
			}
			break
		case reflect.Bool:
			boolEnvValue, err := strconv.ParseBool(envValue)
			if err == nil {
				val.Field(idxNumField).SetBool(boolEnvValue)
			}
			break
		case reflect.Slice:
			splitter := ":"
			if runtime.GOOS == "windows" {
				splitter = ";"
			}
			stringEnvValues := strings.Split(envValue, splitter)
			val.Field(idxNumField).Set(reflect.MakeSlice(val.Field(idxNumField).Type(), len(stringEnvValues), len(stringEnvValues)))
			for idxSlice := 0; idxSlice < len(stringEnvValues); idxSlice++ {
				spew.Dump(val.Field(idxNumField).Index(idxSlice))
				val.Field(idxNumField).Index(idxSlice).SetString(stringEnvValues[idxSlice])
			}
			break
		}
	}

	return obj
}
