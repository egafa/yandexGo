package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env"
	handler "github.com/egafa/yandexGo/api/handler"
	model "github.com/egafa/yandexGo/api/model"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type cfg struct {
	addr           string
	store_Interval int
	store_File     string
	//restore        bool
}

func initconfig() cfg {
	cfg := cfg{}

	cfg.addr = "127.0.0.1:8080"

	cfg.store_Interval = 5

	//cfg.store_File = "\tmp\devops-metrics-db.json"
	cfg.store_File = "devops-metrics-db.json"

	env.Parse(&cfg)
	return cfg
}

func SaveMapMetric(m model.Metric, cfg cfg) {

	for i := 0; i < 25; i++ {
		time.Sleep(time.Duration(cfg.store_Interval) * time.Second)

		file, err := os.Create(cfg.store_File)
		if err != nil {
			log.Fatalf("Ошибка создания файла: %v", err)
			continue
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		err = encoder.Encode(m)
		if err != nil {
			log.Fatalf("Ошибка сериализации: %v", err)
		}

	}
}

func main() {
	cfg := initconfig()

	model.InitMapMetricVal()
	m := model.GetMetricVal()

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/", func(r chi.Router) {
		r.Get("/", handler.ListMetricsChiHandleFunc(m))
	})

	r.Route("/update", func(r chi.Router) {
		r.Post("/{typeMetric}/{nammeMetric}/{valueMetric}", handler.UpdateMetricHandlerChi(m))
		r.Post("/", handler.UpdateMetricHandlerChi(m))
	})

	r.Route("/value", func(r chi.Router) {
		r.Get("/{typeMetric}/{nammeMetric}", handler.ValueMetricHandlerChi(m))
		r.Post("/", handler.ValueMetricHandlerChi(m))
	})

	srv := &http.Server{
		Handler: r,
	}

	srv.Addr = cfg.addr

	go SaveMapMetric(m, cfg)

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
		log.Print("HTTP server Shutdown")
		close(idleConnsClosed)
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}

	<-idleConnsClosed

	log.Print("HTTP server close")

}
