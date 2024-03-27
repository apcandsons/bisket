package internal

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os/exec"
	"syscall"
	"time"
)

type AppInstanceState int

const (
	Pulling AppInstanceState = iota
	Running
	Stopped
)

type AppInstanceEvent struct {
	eventType AppInstanceState
	timestamp time.Time
}

type AppInstance struct {
	Name       string
	Version    string
	Port       string
	RepoUrl    string
	State      AppInstanceState
	RunCommand []string
	Channel    chan error
	CurrCmd    *exec.Cmd
}

func (app *AppInstance) RecordEvent(eventType AppInstanceState) {
	app.State = eventType
	var event = AppInstanceEvent{
		eventType: eventType,
		timestamp: time.Now(),
	}
	slog.Info(fmt.Sprintf("Event: %v at %v", app.State, event.timestamp))
}

func (app *AppInstance) Start() error {
	app.Channel = make(chan error)
	// defer close(app.Channel)

	var appStdWriter = AppStdoutWriter{Name: app.Name, Ver: app.Version}
	var appErrWriter = AppStderrWriter{Name: app.Name, Ver: app.Version}

	app.RecordEvent(Pulling)
	appStdWriter.Info(fmt.Sprintf("Pulling version: %v", app.Version))
	err := CloneOrPullVersion(app.RepoUrl, app.Version)
	if err != nil {
		appErrWriter.Error(fmt.Sprintf("Error pulling version: %v", err))
		return err
	}

	go func() {
		app.RecordEvent(Running)
		appStdWriter.Info(fmt.Sprintf("Running %v/%v on %v", app.Name, app.Version, app.Port))
		appDir := fmt.Sprintf(".bisket/%v", app.Version)

		for _, line := range app.RunCommand {
			appStdWriter.Debug(fmt.Sprintf("Running command: %v", line))
			cmd := exec.Command("sh", "-c", line)
			cmd.Dir = appDir
			cmd.Env = append(cmd.Env, fmt.Sprintf("BISKET_PORT=%s", app.Port))
			cmd.Stdout = &appStdWriter
			cmd.Stderr = &appErrWriter

			app.CurrCmd = cmd
			if err := cmd.Run(); err != nil {
				appErrWriter.Error(fmt.Sprintf("Error starting command: %v", err))
				app.Channel <- err
				return
			}
		}
		app.RecordEvent(Stopped)
		appStdWriter.Info("Server stopped")
		app.Channel <- nil
	}()

	return nil
}

func (app *AppInstance) HandleConnection(w http.ResponseWriter, r *http.Request) {
	targetUrl, _ := url.Parse("http://localhost:" + app.Port)
	proxy := httputil.NewSingleHostReverseProxy(targetUrl)
	proxy.ServeHTTP(w, r)
}

func (app *AppInstance) Stop() {
	err := app.CurrCmd.Process.Signal(syscall.SIGINT)
	if err != nil {
		slog.Warn(fmt.Sprintf("Error stopping app: %v", err))
	}
	app.RecordEvent(Stopped)
}

type AppStdoutWriter struct {
	Name string
	Ver  string
}

func (w *AppStdoutWriter) Write(p []byte) (int, error) {
	w.Info(string(p))
	return len(p), nil
}

func (w *AppStdoutWriter) Info(p string) {
	slog.Info(fmt.Sprintf("[%v/%v] %v", w.Name, w.Ver, p))
}

func (w *AppStdoutWriter) Debug(m string) {
	slog.Debug(fmt.Sprintf("[%v/%v] %v", w.Name, w.Ver, m))
}

type AppStderrWriter struct {
	Name string
	Ver  string
}

func (w *AppStderrWriter) Write(p []byte) (int, error) {
	w.Error(string(p))
	return len(p), nil
}

func (w *AppStderrWriter) Error(p string) {
	slog.Error(fmt.Sprintf("[%v/%v] %v", w.Name, w.Ver, p))
}
