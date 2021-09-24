package server

import (
	"context"
	"net/http"

	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

//go:generate moq -out ./server_generated_mocks_test.go . ClearableCache

type ClearableCache interface {
	Clear()
}

var httpServer *dphttp.Server

func Start(ctx context.Context, cache ClearableCache, bindAdd string, errorChan chan error) {
	router := mux.NewRouter()
	router.Path("/healthcheck").Methods(http.MethodGet).HandlerFunc(HealthCheck)
	router.Path("/dropcache").Methods(http.MethodPost).HandlerFunc(ClearCache(cache))

	httpServer = dphttp.NewServer(bindAdd, router)

	// Disable this here to allow main to manage graceful shutdown of the entire app.
	httpServer.HandleOSSignals = false

	go func() {
		log.Info(ctx, "[HTTPServer]: starting import-reporter HTTP server", log.Data{
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
		ctx := r.Context()
		log.Info(ctx, "dropping import-reporter cache")
		cache.Clear()
		w.WriteHeader(http.StatusOK)
	}
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.Info(ctx, "health check endpoint")
	w.WriteHeader(http.StatusOK)
}

func Shutdown(ctx context.Context) {
	httpServer.Shutdown(ctx)
	log.Info(ctx, "graceful shutdown complete")
}
