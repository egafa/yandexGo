package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"

	handler "github.com/egafa/ytest/api/handler"
	model "github.com/egafa/ytest/api/model"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {

	addr := "127.0.0.1:8080"

	model.InitMapMetricVal()

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/", func(r chi.Router) {
		r.Use()
		r.Get("/", handler.ListMetricsChi)
	})

	r.Route("/update", func(r chi.Router) {
		r.Get("/{typeMetric}/{nammeMetric}/{valueMetric}", handler.UpdateMetricHandlerChi)
	})

	r.Route("/value", func(r chi.Router) {
		r.Get("/{typeMetric}/{nammeMetric}", handler.ValueMetricHandlerChi)
	})

	srv := &http.Server{
		Handler: r,
	}

	srv.Addr = addr

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		// We received an interrupt signal, shut down.
		if err := srv.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}

	<-idleConnsClosed

}
