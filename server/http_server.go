package server

import (
	"context"
	"github.com/ONSdigital/dp-import-reporter/logging"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"net/http"
)

//go:generate moq -out ../mocks/server_generated_mocks.go -pkg mocks . ClearableCache

type ClearableCache interface {
	Clear()
}

var httpServer *server.Server
var logger = logging.Logger{"sever.HTTPServer"}

func Start(cache ClearableCache, bindAdd string, errorChan chan error) {
	router := mux.NewRouter()
	router.Path("/healthcheck").Methods(http.MethodGet).HandlerFunc(HealthCheck)
	router.Path("/dropcache").Methods(http.MethodPost).HandlerFunc(ClearCache(cache))

	httpServer = server.New(bindAdd, router)

	// Disable this here to allow main to manage graceful shutdown of the entire app.
	httpServer.HandleOSSignals = false

	go func() {
		log.Info("[HTTPServer]: starting import-reporter HTTP server", log.Data{
			"/healthcheck": http.MethodGet,
			"/dropcache":   http.MethodPost,
		})
		if err := httpServer.ListenAndServe(); err != nil {
			errorChan <- errors.Wrap(err, "httpServer.ListenAndServe returned error")
		}
	}()
}

func ClearCache(cache ClearableCache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("dropping import-reporter cache", nil)
		cache.Clear()
		w.WriteHeader(http.StatusOK)
	}
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	logger.Info("health check endpoint", nil)
	w.WriteHeader(http.StatusOK)
}

func Shutdown(ctx context.Context) {
	httpServer.Shutdown(ctx)
	logger.Info("graceful shutdown complete", nil)
}
