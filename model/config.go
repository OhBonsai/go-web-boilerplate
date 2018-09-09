package model

type Config struct {
	ServiceSettings       ServiceSettings
	LogSettings           LogSettings
	SqlSettings           SqlSettings
	LocalizationSettings  LocalizationSettings
}

type ServiceSettings struct {
	SiteURL                                           *string
	WebsocketURL                                      *string
	LicenseFileLocation                               *string
	ListenAddress                                     *string
	ConnectionSecurity                                *string
	TLSCertFile                                       *string
	TLSKeyFile                                        *string
	UseLetsEncrypt                                    *bool
	LetsEncryptCertificateCacheFile                   *string
	Forward80To443                                    *bool
	ReadTimeout                                       *int
	WriteTimeout                                      *int
	MaximumLoginAttempts                              *int
	GoroutineHealthThreshold                          *int
	GoogleDeveloperKey                                string
	EnableOAuthServiceProvider                        bool
	EnableIncomingWebhooks                            bool
	EnableOutgoingWebhooks                            bool
	EnableCommands                                    *bool
	EnableOnlyAdminIntegrations                       *bool
	EnablePostUsernameOverride                        bool
	EnablePostIconOverride                            bool
	EnableLinkPreviews                                *bool
	EnableTesting                                     bool
	EnableDeveloper                                   *bool
	EnableSecurityFixAlert                            *bool
	EnableInsecureOutgoingConnections                 *bool
	AllowedUntrustedInternalConnections               *string
	EnableMultifactorAuthentication                   *bool
	EnforceMultifactorAuthentication                  *bool
	EnableUserAccessTokens                            *bool
	AllowCorsFrom                                     *string
	CorsExposedHeaders                                *string
	CorsAllowCredentials                              *bool
	CorsDebug                                         *bool
	AllowCookiesForSubdomains                         *bool
	SessionLengthWebInDays                            *int
	SessionLengthMobileInDays                         *int
	SessionLengthSSOInDays                            *int
	SessionCacheInMinutes                             *int
	SessionIdleTimeoutInMinutes                       *int
	WebsocketSecurePort                               *int
	WebsocketPort                                     *int
	WebserverMode                                     *string
	EnableCustomEmoji                                 *bool
	EnableEmojiPicker                                 *bool
	EnableGifPicker                                   *bool
	GfycatApiKey                                      *string
	GfycatApiSecret                                   *string
	RestrictCustomEmojiCreation                       *string
	RestrictPostDelete                                *string
	AllowEditPost                                     *string
	PostEditTimeLimit                                 *int
	TimeBetweenUserTypingUpdatesMilliseconds          *int64
	EnablePostSearch                                  *bool
	EnableUserTypingMessages                          *bool
	EnableChannelViewedMessages                       *bool
	EnableUserStatuses                                *bool
	ExperimentalEnableAuthenticationTransfer          *bool
	ClusterLogTimeoutMilliseconds                     *int
	CloseUnusedDirectMessages                         *bool
	EnablePreviewFeatures                             *bool
	EnableTutorial                                    *bool
	ExperimentalEnableDefaultChannelLeaveJoinMessages *bool
	ExperimentalGroupUnreadChannels                   *string
	ExperimentalChannelOrganization                   *bool
	ImageProxyType                                    *string
	ImageProxyURL                                     *string
	ImageProxyOptions                                 *string
	EnableAPITeamDeletion                             *bool
	ExperimentalEnableHardenedMode                    *bool
	ExperimentalLimitClientConfig                     *bool
	EnableEmailInvitations                            *bool
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
