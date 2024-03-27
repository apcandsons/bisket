package internal

import (
	"fmt"
	"net/http"
)

type AdminService struct {
	Server   *Server
	Handlers map[string]func(http.ResponseWriter, *http.Request) error
}

func (a *AdminService) Init(server *Server) *AdminService {
	a.Server = server
	a.Handlers = map[string]func(http.ResponseWriter, *http.Request) error{
		"/apps":              a.GetAppInstances,
		"/repo/tags/refresh": a.RefreshRepoTags,
	}
	return a
}

func (a *AdminService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if handler, ok := a.Handlers[r.URL.Path]; ok {
		err := handler(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	http.Error(w, "Not found", http.StatusNotFound)
}

func (a *AdminService) GetAppInstances(w http.ResponseWriter, r *http.Request) error {
	w.WriteHeader(http.StatusOK)
	for _, app := range a.Server.AppInstances {
		_, err := w.Write([]byte(fmt.Sprintf("%s(%s):%s [%d]", app.Name, app.Version, app.Port, app.State)))
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *AdminService) RefreshRepoTags(w http.ResponseWriter, r *http.Request) error {
	err := a.Server.Repo.RefreshTags()
	if err != nil {
		return err
	}
	a.Server.refreshAppInstances()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Tags refreshed"))
	return nil
}
