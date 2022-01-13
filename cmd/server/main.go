package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	handler "github.com/egafa/yandexGo/api/handler"
	model "github.com/egafa/yandexGo/api/model"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {

	addr := "127.0.0.1:8080"

	mapMetric := model.NewMapMetric() //избавился от глобальной переменной

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/", func(r chi.Router) {
		r.Get("/", handler.ListMetricsChiHandleFunc(mapMetric))
	})

	r.Route("/update", func(r chi.Router) {
		r.Post("/{typeMetric}/{nammeMetric}/{valueMetric}", handler.UpdateMetricHandlerChi(mapMetric))
		r.Post("/", handler.UpdateMetricHandlerChi(mapMetric))
	})

	r.Route("/value", func(r chi.Router) {
		r.Get("/{typeMetric}/{nammeMetric}", handler.ValueMetricHandlerChi(mapMetric))
		r.Post("/", handler.ValueMetricHandlerChi(mapMetric))
	})

	srv := &http.Server{
		Handler: r,
		Addr:    addr,
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
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
