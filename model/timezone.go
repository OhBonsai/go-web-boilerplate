package model

import (
	"encoding/json"
	"io"
)

type SupportedTimezones []string

func TimezonesToJson(timezoneList []string) string {
	b, _ := json.Marshal(timezoneList)
	return string(b)
}

func TimezonesFromJson(data io.Reader) SupportedTimezones {
	var timezones SupportedTimezones
	json.NewDecoder(data).Decode(&timezones)
	return timezones
}

func DefaultUserTimezone() map[string]string {
	defaultTimezone := make(map[string]string)
	defaultTimezone["useAutomaticTimezone"] = "true"
	defaultTimezone["automaticTimezone"] = ""
	defaultTimezone["manualTimezone"] = ""

	return defaultTimezone
}

var DefaultSupportedTimezones = []string{
	"Asia/Shanghai",
}