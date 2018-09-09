package app

import (
	"net/http"
	"net"
	"context"
	"github.com/OhBonsai/go-web-boilerplate/store"
	"github.com/OhBonsai/go-web-boilerplate/mlog"
	"github.com/gorilla/mux"
	"time"
)

const TIME_TO_WAIT_FOR_CONNECTIONS_TO_CLOSE_ON_SERVER_SHUTDOWN = time.Second

type Server struct {
	Store			  store.Store
	//WebSocketRouter   *WebSocketRouter
	Router            *mux.Router
	Server            *http.Server
	ListenAddr        *net.TCPAddr
	RateLimiter       *RateLimiter

	didFinishListen   chan struct{}
}


func (a *App) StopServer() {
	if a.Srv.Server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), TIME_TO_WAIT_FOR_CONNECTIONS_TO_CLOSE_ON_SERVER_SHUTDOWN)
		defer cancel()
		didShutdown := false
		for a.Srv.didFinishListen != nil && !didShutdown {
			if err := a.Srv.Server.Shutdown(ctx); err != nil {
				mlog.Warn(err.Error())
			}
			timer := time.NewTimer(time.Millisecond * 50)
			select {
			case <-a.Srv.didFinishListen:
				didShutdown = true
			case <-timer.C:
			}
			timer.Stop()
		}
		a.Srv.Server.Close()
		a.Srv.Server = nil
	}
}