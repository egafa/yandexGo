package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	handler "github.com/egafa/yandexGo/api/handler"
	model "github.com/egafa/yandexGo/api/model"
	"github.com/egafa/yandexGo/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {

	cfg := config.LoadConfigServer()
	log.Println("Запуск Сервера ", cfg.AddrServer)
	log.Println(" файл ", cfg.StoreFile, " интервал сохранения ", cfg.StoreInterval, "флаг восстановления", cfg.Restore)

	mapMetric := model.NewMapMetric()

	if cfg.StoreFile != "" {
		mapMetric.FileName = cfg.StoreFile
		log.Println("Путь к файлу метрик ", mapMetric.FileName)
	}

	if cfg.StoreInterval == 0 {
		mapMetric.SetFlagSave(true)
	}

	if cfg.Restore {
		mapMetric.LoadFromFile()
	}

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
		Addr:    cfg.AddrServer,
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
		log.Print("HTTP server Shutdown")
		close(idleConnsClosed)
	}()

	go SaveToFileTimer(mapMetric, cfg)

	log.Print("Запуск сервера HTTP")

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}

	<-idleConnsClosed

	//SaveMapMetric(mapMetric, cfg)
	log.Print("HTTP server close")

}

func SaveToFileTimer(m model.MapMetric, cfg *config.Config_Server) {
	if cfg.StoreInterval == 0 {
		return
	}
	for {
		time.Sleep(time.Duration(cfg.StoreInterval) * time.Second)
		m.SaveToFile()
	}
}
