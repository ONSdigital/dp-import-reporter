package healthcheck

import (
	"github.com/gorilla/mux"
	"net/http"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"
	"context"
)

//go:generate moq -out ../mocks/healthcheck_generated_mocks.go -pkg mocks . MessageProducer Marshaller

const (
	startingHealthCheckServer = "Starting healthcheck server..."
	listenAndServeErr         = "httpServer.ListenAndServe returned error"
	gracefulShutdownComplete  = "healthcheck server graceful shutdown complete"
)

var httpServer *server.Server

func NewHandler(bindAdd string, errorChan chan error) {
	router := mux.NewRouter()
	router.Path("/healthcheck").HandlerFunc(checkHealth)

	httpServer = server.New(bindAdd, router)
	// Disable this here to allow main to manage graceful shutdown of the entire app.
	httpServer.HandleOSSignals = false

	go func() {
		log.Info(startingHealthCheckServer, nil)
		if err := httpServer.ListenAndServe(); err != nil {
			log.ErrorC(listenAndServeErr, err, nil)
			errorChan <- err
		}
	}()
}

func checkHealth(w http.ResponseWriter, r *http.Request) {
	log.Info("Health check endpoint", nil)
	w.WriteHeader(http.StatusOK)
}

func Shutdown(ctx context.Context) {
	httpServer.Shutdown(ctx)
	log.Info(gracefulShutdownComplete, nil)
}
