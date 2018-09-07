package utils

import (
	"path/filepath"
	"os"
	"go-web-boilerplate/model"
	"go-web-boilerplate/mlog"
	"strings"
	"github.com/fsnotify/fsnotify"
	"io"
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
	"encoding/json"
	"github.com/mattermost/mattermost-server/utils/jsonutils"
	"bytes"
	"reflect"
)

const (
	LOG_FILENAME    = "bonsai.log"
)


type ConfigWatcher struct {
	watcher *fsnotify.Watcher
	close   chan struct{}
	closed  chan struct{}
}


func FindDir(dir string) (string, bool) {
	for _, parent := range []string{".", "..", "../..", "../../.."} {
		foundDir, err := filepath.Abs(filepath.Join(parent, dir))
		if err != nil {
			continue
		} else if _, err := os.Stat(foundDir); err == nil {
			return foundDir, true
		}
	}
	return "./", false
}


func GetLogFileLocation(fileLocation string) string {
	if fileLocation == "" {
		fileLocation, _ = FindDir("logs")
	}

	return filepath.Join(fileLocation, LOG_FILENAME)
}

func MloggerConfigFromLoggerConfig(s *model.LogSettings) *mlog.LoggerConfiguration {
	return &mlog.LoggerConfiguration{
		EnableConsole: s.EnableConsole,
		ConsoleJson:   *s.ConsoleJson,
		ConsoleLevel:  strings.ToLower(s.ConsoleLevel),
		EnableFile:    s.EnableFile,
		FileJson:      *s.FileJson,
		FileLevel:     strings.ToLower(s.FileLevel),
		FileLocation:  GetLogFileLocation(s.FileLocation),
	}
}

func FindConfigFile(fileName string) (path string) {
	if filepath.IsAbs(fileName) {
		if _, err := os.Stat(fileName); err == nil {
			return fileName
		}
	} else {
		for _, dir := range []string{"./config", "../config", "../../config", "../../../config", "."} {
			path, _ := filepath.Abs(filepath.Join(dir, fileName))
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}
	return ""
}

func ReadConfig(r io.Reader, allowEnvironmentOverrides bool) (*model.Config, map[string]interface{}, error) {
	// Pre-flight check the syntax of the configuration file to improve error messaging.
	configData, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, nil, err
	} else {
		var rawConfig interface{}
		if err := json.Unmarshal(configData, &rawConfig); err != nil {
			return nil, nil, jsonutils.HumanizeJsonError(err, configData)
		}
	}

	v := newViper(allowEnvironmentOverrides)
	if err := v.ReadConfig(bytes.NewReader(configData)); err != nil {
		return nil, nil, err
	}

	var config model.Config
	unmarshalErr := v.Unmarshal(&config)
	envConfig := v.EnvSettings()

	var envErr error
	if envConfig, envErr = fixEnvSettingsCase(envConfig); envErr != nil {
		return nil, nil, envErr
	}

	return &config, envConfig, unmarshalErr
}

func fixEnvSettingsCase(in map[string]interface{}) (out map[string]interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			mlog.Error(fmt.Sprintf("Panicked in fixEnvSettingsCase. This should never happen. %v", r))
			out = in
		}
	}()

	var fixCase func(map[string]interface{}, reflect.Type) map[string]interface{}
	fixCase = func(in map[string]interface{}, t reflect.Type) map[string]interface{} {
		if t.Kind() != reflect.Struct {
			// Should never hit this, but this will prevent a panic if that does happen somehow
			return nil
		}

		out := make(map[string]interface{}, len(in))

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)

			key := field.Name
			if value, ok := in[strings.ToLower(key)]; ok {
				if valueAsMap, ok := value.(map[string]interface{}); ok {
					out[key] = fixCase(valueAsMap, field.Type)
				} else {
					out[key] = value
				}
			}
		}

		return out
	}

	out = fixCase(in, reflect.TypeOf(model.Config{}))

	return
}


func ReadConfigFile(path string, allowEnvironmentOverrides bool) (*model.Config, map[string]interface{}, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	return ReadConfig(f, allowEnvironmentOverrides)
}

func EnsureConfigFile(fileName string) (string, error) {
	if configFile := FindConfigFile(fileName); configFile != "" {
		return configFile, nil
	}
	if defaultPath := FindConfigFile("default.json"); defaultPath != "" {
		destPath := filepath.Join(filepath.Dir(defaultPath), fileName)
		src, err := os.Open(defaultPath)
		if err != nil {
			return "", err
		}
		defer src.Close()
		dest, err := os.OpenFile(destPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return "", err
		}
		defer dest.Close()
		if _, err := io.Copy(dest, src); err == nil {
			return destPath, nil
		}
	}
	return "", fmt.Errorf("no config file found")
}


func LoadConfig(fileName string) (*model.Config, string, map[string]interface{}, *model.AppError) {
	var configPath string

	if fileName != filepath.Base(fileName) {
		configPath = fileName
	} else {
		if path, err := EnsureConfigFile(fileName); err != nil {
			appErr := model.NewAppError("LoadConfig", "utils.config.load_config.opening.panic", map[string]interface{}{"Filename": fileName, "Error": err.Error()}, "", 0)
			return nil, "", nil, appErr
		} else {
			configPath = path
		}
	}

	config, envConfig, err := ReadConfigFile(configPath, true)
	if err != nil {
		appErr := model.NewAppError("LoadConfig", "utils.config.load_config.decoding.panic", map[string]interface{}{"Filename": fileName, "Error": err.Error()}, "", 0)
		return nil, "", nil, appErr
	}

	needSave := len(config.SqlSettings.AtRestEncryptKey) == 0 || len(*config.FileSettings.PublicLinkSalt) == 0 ||
		len(config.EmailSettings.InviteSalt) == 0

	config.SetDefaults()

	if err := config.IsValid(); err != nil {
		return nil, "", nil, err
	}

	if needSave {
		if err := SaveConfig(configPath, config); err != nil {
			mlog.Warn(err.Error())
		}
	}

	if err := ValidateLocales(config); err != nil {
		if err := SaveConfig(configPath, config); err != nil {
			mlog.Warn(err.Error())
		}
	}

	if *config.FileSettings.DriverName == model.IMAGE_DRIVER_LOCAL {
		dir := config.FileSettings.Directory
		if len(dir) > 0 && dir[len(dir)-1:] != "/" {
			config.FileSettings.Directory += "/"
		}
	}

	return config, configPath, envConfig, nil
}

func newViper(allowEnvironmentOverrides bool) *viper.Viper {
	v := viper.New()

	v.SetConfigType("json")

	if allowEnvironmentOverrides {
		v.SetEnvPrefix("mm")
		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		v.AutomaticEnv()
	}

	// Set zeroed defaults for all the config settings so that Viper knows what environment variables
	// it needs to be looking for. The correct defaults will later be applied using Config.SetDefaults.
	defaults := getDefaultsFromStruct(model.Config{})

	for key, value := range defaults {
		v.SetDefault(key, value)
	}

	return v
}


func getDefaultsFromStruct(s interface{}) map[string]interface{} {
	return flattenStructToMap(structToMap(reflect.TypeOf(s)))
}

func flattenStructToMap(in map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{})

	for key, value := range in {
		if valueAsMap, ok := value.(map[string]interface{}); ok {
			sub := flattenStructToMap(valueAsMap)

			for subKey, subValue := range sub {
				out[key+"."+subKey] = subValue
			}
		} else {
			out[key] = value
		}
	}

	return out
}

func structToMap(t reflect.Type) (out map[string]interface{}) {
	defer func() {
		if r := recover(); r != nil {
			mlog.Error(fmt.Sprintf("Panicked in structToMap. This should never happen. %v", r))
		}
	}()

	if t.Kind() != reflect.Struct {
		// Should never hit this, but this will prevent a panic if that does happen somehow
		return nil
	}

	out = map[string]interface{}{}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		var value interface{}

		switch field.Type.Kind() {
		case reflect.Struct:
			value = structToMap(field.Type)
		case reflect.Ptr:
			indirectType := field.Type.Elem()

			if indirectType.Kind() == reflect.Struct {
				// Follow pointers to structs since we need to define defaults for their fields
				value = structToMap(indirectType)
			} else {
				value = nil
			}
		default:
			value = reflect.Zero(field.Type).Interface()
		}

		out[field.Name] = value
	}

	return
}

func (o *Config) SetDefaults() {
	o.LdapSettings.SetDefaults()
	o.SamlSettings.SetDefaults()

	if o.TeamSettings.TeammateNameDisplay == nil {
		o.TeamSettings.TeammateNameDisplay = NewString(SHOW_USERNAME)

		if *o.SamlSettings.Enable || *o.LdapSettings.Enable {
			*o.TeamSettings.TeammateNameDisplay = SHOW_FULLNAME
		}
	}

	o.SqlSettings.SetDefaults()
	o.FileSettings.SetDefaults()
	o.EmailSettings.SetDefaults()
	o.ServiceSettings.SetDefaults()
	o.PasswordSettings.SetDefaults()
	o.TeamSettings.SetDefaults()
	o.MetricsSettings.SetDefaults()
	o.SupportSettings.SetDefaults()
	o.AnnouncementSettings.SetDefaults()
	o.ThemeSettings.SetDefaults()
	o.ClusterSettings.SetDefaults()
	o.PluginSettings.SetDefaults()
	o.AnalyticsSettings.SetDefaults()
	o.ComplianceSettings.SetDefaults()
	o.LocalizationSettings.SetDefaults()
	o.ElasticsearchSettings.SetDefaults()
	o.NativeAppSettings.SetDefaults()
	o.DataRetentionSettings.SetDefaults()
	o.RateLimitSettings.SetDefaults()
	o.LogSettings.SetDefaults()
	o.JobSettings.SetDefaults()
	o.WebrtcSettings.SetDefaults()
	o.MessageExportSettings.SetDefaults()
	o.TimezoneSettings.SetDefaults()
	o.DisplaySettings.SetDefaults()
}