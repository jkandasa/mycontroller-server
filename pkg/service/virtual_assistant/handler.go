package service

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

const (
	API_PATH = "/api/bot/virtual/assistant/"
)

// virtual assistant will be registered and unregistered dynamically
func RegisterVirtualAssistantServiceRoutes(router *mux.Router) {
	router.PathPrefix(API_PATH).HandlerFunc(handleRoute)
}

func handleRoute(w http.ResponseWriter, r *http.Request) {
	// get assistant id
	path := strings.TrimPrefix(r.URL.Path, API_PATH)
	paths := strings.SplitN(path, "/", 2)

	assistantID := ""
	if len(paths) == 1 {
		r.URL.Path = ""
		assistantID = paths[0]
	} else if len(paths) == 2 {
		assistantID = paths[0]
		r.URL.Path = paths[1]
	}

	if assistantID == "" {
		http.Error(w, "assistant id can not be empty", http.StatusBadRequest)
		return
	}

	assistant := vaService.Get(assistantID)
	if assistant == nil {
		http.Error(w, "requested assistant not available", http.StatusNotFound)
		return
	}

	assistant.ServeHTTP(w, r)
}
