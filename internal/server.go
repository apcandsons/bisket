package internal

import (
	"fmt"
	"log"
	"log/slog"
	"net"
)

type Server struct {
	repo *Repository
	conf *Config
	apps []AppInstance
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
	svr.apps = append(svr.apps, *app)
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

	var port = svr.conf.Port
	// Start the server
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Printf("Server listening on port %v\n", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("Error accepting connection: %v", err)
		}
		go svr.HandleConnection(conn)
	}
}

func (svr *Server) HandleConnection(conn net.Conn) {
	fmt.Printf("Received connection from %v\n", conn.RemoteAddr())
	err := conn.Close()
	if err != nil {
		log.Fatalf("Error closing connection: %v", err)
		return
	}
}
