package internal

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
)

type Server struct {
	Repo            *Repository
	Conf            *Config
	AppInstances    []*AppInstance
	LastAppInstance *AppInstance
}

type HttpServer struct {
	HttpHandler func(http.ResponseWriter, *http.Request)
}

func (h HttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.HttpHandler(w, r)
}

func (svr *Server) Init(cfg *Config, repo *Repository) {
	svr.Repo = repo
	svr.Conf = cfg
	repo.OnVersionUpdate(svr.OnVersionUpdate)
	repo.OnVersionDestroy(svr.OnVersionDestroy)
}

func (svr *Server) OnVersionUpdate(repo *Repository, event *VersionUpdateEvent) error {
	slog.Info(fmt.Sprintf("Version updated to %v", event.latestVersion))
	app, err := svr.CreateAppInstance(event.latestVersion)
	if err != nil {
		return err
	}
	// Destroy the old app instance
	svr.stopOldAppInstances(event.latestVersion)
	svr.AppInstances = append(svr.AppInstances, app)
	return nil
}

func (svr *Server) stopOldAppInstances(currVer string) {
	for _, app := range svr.AppInstances {
		if app.Version != currVer && app.State == Running {
			app.Stop()
		}
	}
}

func (svr *Server) OnVersionDestroy(repo *Repository, event *VersionDestroyEvent) error {
	slog.Info("Version updated to %v\n", event.destroyedVersion)
	// svr.CreateAppInstance(event.destroyedVersion)
	return nil
}

func (svr *Server) CreateAppInstance(version string) (*AppInstance, error) {
	port, err := getFreePort()
	if err != nil {
		slog.Error(fmt.Sprintf("No open port available on this server: %v", err))
		return nil, err
	}

	var app = AppInstance{
		Name:       svr.Repo.AppName,
		Port:       strconv.Itoa(port),
		Repo:       svr.Repo,
		Version:    version,
		RunCommand: svr.Conf.RunCommand,
	}

	if err := app.Start(); err != nil {
		return nil, err
	}

	return &app, nil
}

func getFreePort() (int, error) {
	var a *net.TCPAddr
	a, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		slog.Error(fmt.Sprintf("Error resolving tcp address: %v", err))
		return 0, err
	}
	var l *net.TCPListener
	l, err = net.ListenTCP("tcp", a)
	if err != nil {
		slog.Error(fmt.Sprintf("Error listening on tcp address: %v", err))
		return 0, err
	}
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			slog.Error("Error closing listener: %v", err)
		}
	}(l)
	return l.Addr().(*net.TCPAddr).Port, nil
}

func (svr *Server) Start() error {
	err := svr.Repo.Init(svr.Conf.RepoConfig)
	if err != nil {
		return err
	}

	ch := make(chan error, 2)
	// Run the service server
	go svr.startProxyServer(ch)
	go svr.startAdminServer(ch)
	return <-ch
}

func (svr *Server) startProxyServer(ch chan error) {
	h := HttpServer{HttpHandler: func(w http.ResponseWriter, r *http.Request) {
		if (svr.LastAppInstance != nil) && (svr.LastAppInstance.State == Running) {
			svr.LastAppInstance.HandleConnection(w, r)
			return
		}
		for _, app := range svr.AppInstances {
			if app.State == Running {
				app.HandleConnection(w, r)
				svr.LastAppInstance = app
				return
			}
		}
		slog.Error("No running applications found!")
	}}

	server := http.Server{Addr: ":" + svr.Conf.Port, Handler: h}
	slog.Info("starting proxy server on port: " + svr.Conf.Port)
	if err := server.ListenAndServe(); err != nil {
		ch <- err
		return
	}
	ch <- nil
}

func (svr *Server) startAdminServer(ch chan error) {
	h := AdminService{}
	h.Init(svr)
	server := http.Server{Addr: ":" + svr.Conf.AdminPort, Handler: &h}
	slog.Info("starting admin server on port: " + svr.Conf.AdminPort)
	if err := server.ListenAndServe(); err != nil {
		ch <- err
		return
	}
	ch <- nil
}
