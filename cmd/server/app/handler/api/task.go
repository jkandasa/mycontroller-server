package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	handlerUtils "github.com/mycontroller-org/server/v2/cmd/server/app/handler/utils"
	taskAPI "github.com/mycontroller-org/server/v2/pkg/api/task"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	taskTY "github.com/mycontroller-org/server/v2/pkg/types/task"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

// RegisterTaskRoutes registers task api
func RegisterTaskRoutes(router *mux.Router) {
	router.HandleFunc("/api/task", listTasks).Methods(http.MethodGet)
	router.HandleFunc("/api/task/{id}", getTask).Methods(http.MethodGet)
	router.HandleFunc("/api/task", updateTask).Methods(http.MethodPost)
	router.HandleFunc("/api/task/enable", enableTask).Methods(http.MethodPost)
	router.HandleFunc("/api/task/disable", disableTask).Methods(http.MethodPost)
	router.HandleFunc("/api/task", deleteTasks).Methods(http.MethodDelete)
}

func listTasks(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindMany(w, r, types.EntityTask, &[]taskTY.Config{})
}

func getTask(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindOne(w, r, types.EntityTask, &taskTY.Config{})
}

func updateTask(w http.ResponseWriter, r *http.Request) {
	entity := &taskTY.Config{}
	err := handlerUtils.LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if entity.ID == "" {
		http.Error(w, "id should not be an empty", 400)
		return
	}
	err = taskAPI.SaveAndReload(entity)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func deleteTasks(w http.ResponseWriter, r *http.Request) {
	IDs := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(IDs) > 0 {
			count, err := taskAPI.Delete(IDs)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("deleted: %d", count), nil
		}
		return nil, errors.New("supply id(s)")
	}
	handlerUtils.UpdateData(w, r, &IDs, updateFn)
}

func enableTask(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			err := taskAPI.Enable(ids)
			if err != nil {
				return nil, err
			}
			return "Enabled", nil
		}
		return nil, errors.New("supply a task id")
	}
	handlerUtils.UpdateData(w, r, &ids, updateFn)
}

func disableTask(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			err := taskAPI.Disable(ids)
			if err != nil {
				return nil, err
			}
			return "Disabled", nil
		}
		return nil, errors.New("supply a task id")
	}
	handlerUtils.UpdateData(w, r, &ids, updateFn)
}
