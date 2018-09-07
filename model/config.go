package model

type Config struct {
	LogSettings           LogSettings
	SqlSettings           SqlSettings
	LocalizationSettings  LocalizationSettings
}

type LogSettings struct {
	EnableConsole          bool
	ConsoleLevel           string
	ConsoleJson            *bool
	EnableFile             bool
	FileLevel              string
	FileJson               *bool
	FileLocation           string
	EnableWebhookDebugging bool
	EnableDiagnostics      *bool
}

type LocalizationSettings struct {
	DefaultServerLocale *string
	DefaultClientLocale *string
	AvailableLocales    *string
}

type SqlSettings struct {
	DriverName               *string
	DataSource               *string
	DataSourceReplicas       []string
	DataSourceSearchReplicas []string
	MaxIdleConns             *int
	MaxOpenConns             *int
	Trace                    bool
	AtRestEncryptKey         string
	QueryTimeout             *int
}
