package server

import (
	"github.com/gorilla/mux"
	"net/http"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"
	"context"
)

//go:generate moq -out ../mocks/server_generated_mocks.go -pkg mocks . ClearableCache

const (
	startingServer           = "starting import-reporter HTTP server"
	listenAndServeErr        = "httpServer.ListenAndServe returned error"
	gracefulShutdownComplete = "http server graceful shutdown complete"
	healthCheckPath          = "/healthcheck"
	dropCachePath            = "/dropcache"
)

type ClearableCache interface {
	Clear()
}

var httpServer *server.Server

func Start(cache ClearableCache, bindAdd string, errorChan chan error) {
	router := mux.NewRouter()
	router.Path(healthCheckPath).Methods(http.MethodGet).HandlerFunc(HealthCheck)
	router.Path(dropCachePath).Methods(http.MethodPost).HandlerFunc(ClearCache(cache))

	httpServer = server.New(bindAdd, router)

	// Disable this here to allow main to manage graceful shutdown of the entire app.
	httpServer.HandleOSSignals = false

	go func() {
		log.Info(startingServer, log.Data{
			healthCheckPath: http.MethodGet,
			dropCachePath:   http.MethodPost,
		})
		if err := httpServer.ListenAndServe(); err != nil {
			log.ErrorC(listenAndServeErr, err, nil)
			errorChan <- err
		}
	}()
}

func ClearCache(cache ClearableCache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("Dropping import-reporter cache", nil)
		cache.Clear()
		w.WriteHeader(http.StatusOK)
	}
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	log.Info("Health check endpoint", nil)
	w.WriteHeader(http.StatusOK)
}

func Shutdown(ctx context.Context) {
	httpServer.Shutdown(ctx)
	log.Info(gracefulShutdownComplete, nil)
}
