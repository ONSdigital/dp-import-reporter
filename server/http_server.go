package server

import (
	"context"
	"net/http"

	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

//go:generate moq -out ../mocks/server_generated_mocks.go -pkg mocks . ClearableCache

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
		log.Event(ctx, "[HTTPServer]: starting import-reporter HTTP server", log.INFO, log.Data{
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
		log.Event(ctx, "dropping import-reporter cache", log.INFO)
		cache.Clear()
		w.WriteHeader(http.StatusOK)
	}
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.Event(ctx, "health check endpoint", log.INFO)
	w.WriteHeader(http.StatusOK)
}

func Shutdown(ctx context.Context) {
	httpServer.Shutdown(ctx)
	log.Event(ctx, "graceful shutdown complete", log.INFO)
}
