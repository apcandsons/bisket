package internal

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
)

type Server struct {
	Repo            Repository
	Conf            Config
	AppInstances    []*AppInstance
	LastAppInstance *AppInstance
}

type HttpServer struct {
	HttpHandler func(http.ResponseWriter, *http.Request)
}

func (h HttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.HttpHandler(w, r)
}

func (svr *Server) initRepo() error {
	svr.Repo = Repository{}
	err := svr.Repo.Init(&svr.Conf.RepoConfig)
	if err != nil {
		return err
	}
	svr.Repo.OnVersionUpdate(svr.OnVersionUpdate)
	svr.Repo.OnVersionDestroy(svr.OnVersionDestroy)
	return nil
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
		RepoUrl:    svr.Repo.RepoUrl,
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
	if err := svr.initRepo(); err != nil {
		return err
	}

	svr.refreshAppInstances()

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

type Action int

const (
	DoNothing Action = iota
	CreateInstance
	DestroyInstance
)

type AppInstanceAction struct {
	Version      string
	DesiredState AppInstanceState
	CurrentState AppInstanceState
	Action       Action
}

func (svr *Server) refreshAppInstances() {
	actions := svr.analyzeActions()
	for _, action := range *actions {
		slog.Info(fmt.Sprintf("Version: %v, Desired state: %s, Current state: %s, Action: %s", action.Version, action.DesiredState, action.CurrentState, action.Action))
		if action.Action == CreateInstance {
			app, err := svr.CreateAppInstance(action.Version)
			if err != nil {
				slog.Error(fmt.Sprintf("Error creating app instance: %v", err))
				continue
			}
			svr.AppInstances = append(svr.AppInstances, app)
		} else if action.Action == DestroyInstance {
			app, err := svr.FindAppInstanceByVersion(action.Version)
			if err == nil {
				app.Stop()
				svr.AppInstances = removeAppInstance(svr.AppInstances, app)
			}
		}
	}
}

func removeAppInstance(instances []*AppInstance, app *AppInstance) []*AppInstance {
	var newInstances = make([]*AppInstance, 0)
	for _, instance := range instances {
		if instance != app {
			newInstances = append(newInstances, instance)
		}
	}
	return newInstances
}

func (svr *Server) analyzeActions() *[]AppInstanceAction {
	// First create a list of versions by iterating over the versions in the repository and the versions in the app instances
	var versions = make([]string, 0)
	for _, version := range svr.Repo.Vers {
		versions = append(versions, version.Tag)
	}
	for _, app := range svr.AppInstances {
		versions = append(versions, app.Version)
	}

	// Create actions needed to bring the app instances to the desired state
	var actions = make([]AppInstanceAction, 0)
	for _, version := range versions {
		var action = AppInstanceAction{
			Version:      version,
			DesiredState: Stopped,
			CurrentState: Stopped,
			Action:       DoNothing,
		}

		if _, err := svr.Repo.FindVersionByTag(version); err == nil {
			action.DesiredState = Running
		}

		if app, err := svr.FindAppInstanceByVersion(version); err == nil {
			action.CurrentState = app.State
		}

		if action.CurrentState != action.DesiredState {
			if action.DesiredState == Running {
				action.Action = CreateInstance
			} else {
				action.Action = DestroyInstance
			}
		}
		actions = append(actions, action)
	}
	return &actions
}

func (svr *Server) FindAppInstanceByVersion(version string) (*AppInstance, error) {
	for _, app := range svr.AppInstances {
		if app.Version == version {
			return app, nil
		}
	}
	return nil, fmt.Errorf("AppInstance not found")

}
