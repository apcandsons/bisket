package internal

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
)

type Server struct {
	repo *Repository
	conf *Config
	apps []*AppInstance
}

func (svr *Server) Init(cfg *Config, repo *Repository) {
	svr.repo = repo
	svr.conf = cfg
	repo.OnVersionUpdate(svr.OnVersionUpdate)
	repo.OnVersionDestroy(svr.OnVersionDestroy)
}

func (svr *Server) OnVersionUpdate(repo *Repository, event *VersionUpdateEvent) error {
	slog.Info(fmt.Sprintf("Version updated to %v", event.latestVersion))
	app, err := svr.CreateAppInstance(event.latestVersion)
	if err != nil {
		return err
	}
	svr.apps = append(svr.apps, app)
	return nil
}

func (svr *Server) OnVersionDestroy(repo *Repository, event *VersionDestroyEvent) error {
	slog.Info("Version updated to %v\n", event.destroyedVersion)
	// svr.CreateAppInstance(event.destroyedVersion)
	return nil
}

func (svr *Server) CreateAppInstance(version string) (*AppInstance, error) {
	var app = AppInstance{
		Name:       svr.repo.AppName,
		Port:       "8081",
		Repo:       svr.repo,
		Version:    version,
		RunCommand: svr.conf.RunCommand,
	}

	if err := app.Start(); err != nil {
		return nil, err
	}

	return &app, nil
}

func (svr *Server) Start() error {
	err := svr.repo.Init(svr.conf.RepoConfig)
	if err != nil {
		return err
	}

	http.HandleFunc("/", svr.HandleRequest)

	// Start the server
	fmt.Printf("Server listening on port %v\n", svr.conf.Port)
	if err := http.ListenAndServe(":"+svr.conf.Port, nil); err != nil {
		log.Fatalf("Error accepting connection: %v", err)
		return err
	}
	return nil
}

func (svr *Server) HandleRequest(w http.ResponseWriter, r *http.Request) {
	// Proxy the request to the app instance
	for _, app := range svr.apps {
		if app.State == Running {
			app.HandleConnection(w, r)
			return
		}
	}
	slog.Error("No running applications found!")
}
