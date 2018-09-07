package main

import (
	"os"

	"github.com/spf13/cobra"

	"go-web-boilerplate/app"
	"github.com/mattermost/mattermost-server/mlog"
	"github.com/mattermost/mattermost-server/utils"
	"net/url"
	"github.com/mattermost/mattermost-server/model"
	"fmt"
	"github.com/mattermost/mattermost-server/api4"
	"github.com/mattermost/mattermost-server/wsapi"
	"github.com/mattermost/mattermost-server/web"
)

const (
	SESSIONS_CLEANUP_BATCH_SIZE = 1000
)

var serverCmd = &cobra.Command{
	Use:          "server",
	Short:        "Run the server",
	RunE:         serverCmdF,
	SilenceUsage: true,
}

func init(){
	RootCmd.AddCommand(serverCmd)
	RootCmd.RunE = serverCmdF
}

func serverCmdF(command *cobra.Command, args []string) error {
	config, err := command.Flags().GetString("config")

	if err != nil {
		return err
	}

	watch, _ := command.Flags().GetBool("watch")
	interruptChan := make(chan os.Signal, 1)
	return runServer(config, watch, interruptChan)
}

func runServer(configFileLoc string, watch bool, interruptChan chan os.Signal) error {
	options := []app.Option{app.ConfigFile(configFileLoc)}

	if !watch{
		options = append(options, app.DisableConfigWatch)
	}

	a, err := app.New(options...)
	if err != nil {
		mlog.Critical(err.Error())
		return err
	}

	defer a.Shutdown()

	utils.TestConnection(a.Config())

	pwd, _ := os.Getwd()

	if _, err := url.ParseRequestURI(*a.Config().ServiceSettings.SiteURL); err != nil {
		mlog.Error("SiteURL must be set. Some features will operate incorrectly if the SiteURL is not set. See documentation for details: http://about.mattermost.com/default-site-url")
	}


	mlog.Info(fmt.Sprintf("Current version is %v (%v/%v/%v/%v)", model.CurrentVersion, model.BuildNumber, model.BuildDate, model.BuildHash, model.BuildHashEnterprise))
	mlog.Info(fmt.Sprintf("Enterprise Enabled: %v", model.BuildEnterpriseReady))
	mlog.Info(fmt.Sprintf("Current working directory is %v", pwd))
	mlog.Info(fmt.Sprintf("Loaded config file from %v", utils.FindConfigFile(configFileLocation)))

	backend, appErr := a.FileBackend()
	if appErr == nil {
		appErr = backend.TestConnection()
	}
	if appErr != nil {
		mlog.Error("Problem with file storage settings: " + appErr.Error())
	}

	a.AddConfigListener(func(prevCfg, cfg *model.Config) {
		if *cfg.PluginSettings.Enable {
			a.InitPlugins(*cfg.PluginSettings.Directory, *a.Config().PluginSettings.ClientDirectory, nil)
		} else {
			a.ShutDownPlugins()
		}
	})

	serverErr := a.StartServer()
	if serverErr != nil {
		mlog.Critical(serverErr.Error())
		return serverErr
	}

	api := api4.Init(a, a.Srv.Router)
	wsapi.Init(a, a.Srv.WebSocketRouter)
	web.NewWeb(a, a.Srv.Router)

	a.ReloadConfig()
}