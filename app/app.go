package app

import (
	"go-web-boilerplate/mlog"
	"sync/atomic"
	"go-web-boilerplate/model"
	"github.com/gorilla/mux"
	"go-web-boilerplate/utils"
	"go-web-boilerplate/store"
	"net/http"
	"fmt"
	"crypto/ecdsa"
	"github.com/mattermost/mattermost-server/einterfaces"
	"go-web-boilerplate/store/sqlstore"
)

type App struct {
	goroutineCount          int32
	goroutineExitSignal     chan struct{}

	Srv                     *Server
	Log                     *mlog.Logger
	newStore 				func() store.Store
	sessionCache         	*utils.Cache

	Cluster          		einterfaces.ClusterInterface
	Metrics          		einterfaces.MetricsInterface


	config                  atomic.Value
	envConfig               map[string]interface{}
	configFile              string
	siteURL                 string
	configListeners         map[string]func(*model.Config, *model.Config)
	configListenerId     	string
	logListenerId        	string
	disableConfigWatch		bool
	configWatcher        	*utils.ConfigWatcher

	timezones 				atomic.Value

	asymmetricSigningKey    *ecdsa.PrivateKey
}

var appCount = 0


func New(options ...Option) (outApp *App, outErr error) {
	appCount ++
	ensureJustOneAppAlive()
	app := simpleInitApp()

	// decorator for app __new__
	defer func() {
		if outErr != nil {app.Shutdown()}
	}()

	for _, option := range options {option(app)}

	// load config
	if err := app.LoadConfig(app.configFile); err != nil {
		return nil, err
	}

	app.addLogger().
		addConfigWatcher().
		addTimeZoneSupport().
		addI18nSupport().
		addStore().
		addBuiltInPlugins().
		addRoute().
		addWebSocket()

	return app, nil

}

func ensureJustOneAppAlive() {
	if appCount > 1 {
		panic("Only one App should exist at a time. Did you forget to call Shutdown()?")
	}
}

func simpleInitApp() *App {
	return &App{
		goroutineExitSignal: make(chan struct{}, 1),
		Srv:                 &Server{Router: mux.NewRouter()},
		sessionCache:     	 utils.NewLru(model.SESSION_CACHE_SIZE),
		configFile:          "./config/config.json",
		configListeners:     make(map[string]func(*model.Config, *model.Config)),
	}
}

func (a *App) addLogger() *App{
	a.Log = mlog.NewLogger(utils.MloggerConfigFromLoggerConfig(&a.Config().LogSettings))
	mlog.RedirectStdLog(a.Log)
	mlog.InitGlobalLogger(a.Log)
	return a
}

func (a *App) addConfigWatcher() *App{
	return a
}

func (a *App) addTimeZoneSupport() *App{
	return a
}

func (a *App) addI18nSupport() *App{
	return a
}

func (a *App) addStore() *App{
	if a.newStore == nil {
		a.newStore = func() store.Store {
			return store.NewLayeredStore(sqlstore.NewSqlSupplier(a.Config().SqlSettings, a.Metrics), a.Metrics, a.Cluster)
		}
	}
	a.Srv.Store = a.newStore()
	return a
}
func (a *App) addBuiltInPlugins() *App{
	return a
}

func (a *App) addRoute() *App{
	a.Srv.Router.NotFoundHandler = http.HandlerFunc(a.Handle404)
	return a
}
func (a *App) addWebSocket() *App{
	return a
}


func (a *App) Shutdown() {
	appCount--

	mlog.Info("Stopping Server...")

	a.StopServer()

	if a.Srv.Store != nil {
		a.Srv.Store.Close()
	}
	a.Srv = nil
	mlog.Info("Server stopped")

}


func (a *App) Handle404(w http.ResponseWriter, r *http.Request) {
	err := model.NewAppError("Handle404", "api.context.404.app_error", nil, "", http.StatusNotFound)
	mlog.Debug(fmt.Sprintf("%v: code=404 ip=%v", r.URL.Path, utils.GetIpAddress(r)))
	utils.RenderWebAppError(w, r, err, a.asymmetricSigningKey)
}
