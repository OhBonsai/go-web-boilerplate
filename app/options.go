package app

import "github.com/OhBonsai/go-web-boilerplate/store"

type Option func(a *App)


func StoreOverride(override interface{}) Option {
	return func(a *App) {
		switch o := override.(type) {
		case store.Store:
			a.newStore = func() store.Store {
				return o
			}
		case func(*App) store.Store:
			a.newStore = func() store.Store {
				return o(a)
			}
		default:
			panic("invalid StoreOverride")
		}
	}
}

func ConfigFile(file string) Option {
	return func(a *App) {
		a.configFile = file
	}
}

func DisableConfigWatch(a *App) {
	a.disableConfigWatch = true
}

