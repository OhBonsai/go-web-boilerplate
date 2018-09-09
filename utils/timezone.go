package utils

import (
	"io/ioutil"
	"encoding/json"
	"github.com/OhBonsai/go-web-boilerplate/model"
)

func LoadTimezones(fileName string) model.SupportedTimezones {
	var supportedTimezones model.SupportedTimezones

	if timezoneFile := FindConfigFile(fileName); timezoneFile == "" {
		return model.DefaultSupportedTimezones
	} else if raw, err := ioutil.ReadFile(timezoneFile); err != nil {
		return model.DefaultSupportedTimezones
	} else if err := json.Unmarshal(raw, &supportedTimezones); err != nil {
		return model.DefaultSupportedTimezones
	} else {
		return supportedTimezones
	}
}
