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
	"github.com/egafa/yandexGo/zipcompess"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v4"
)

func main() {

	cfg := config.LoadConfigServer()
	log.Println("Запуск Сервера ", cfg.AddrServer)
	log.Println(" файл ", cfg.StoreFile, " интервал сохранения ", cfg.StoreInterval, "флаг восстановления", cfg.Restore, " Каталог шаблонов ", cfg.TemplateDir, " Key ", cfg.Key)
	log.Println(" databse url ", cfg.DatabaseDSN)

	mapMetric := model.NewMapMetricCongig(cfg)

	//cfg.DatabaseDSN = "postgres://postgres:qwertyd@localhost:5432/exam1"

	ctx, cancel := context.WithCancel(context.Background())

	conn, err := pgx.Connect(ctx, cfg.DatabaseDSN)
	if err != nil {
		log.Println("database error ", err.Error())
		os.Exit(1)
	}
	defer conn.Close(ctx)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(zipcompess.GzipHandle)

	r.Route("/", func(r chi.Router) {
		r.Get("/", handler.ListMetricsChiHandleFunc(mapMetric, cfg))
	})

	r.Route("/update", func(r chi.Router) {
		r.Post("/{typeMetric}/{nammeMetric}/{valueMetric}", handler.UpdateMetricHandlerChi(mapMetric, cfg))
		r.Post("/", handler.UpdateMetricHandlerChi(mapMetric, cfg))
	})

	r.Route("/value", func(r chi.Router) {
		r.Get("/{typeMetric}/{nammeMetric}", handler.ValueMetricHandlerChi(mapMetric, cfg))
		r.Post("/", handler.ValueMetricHandlerChi(mapMetric, cfg))
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
		if err := srv.Shutdown(ctx); err != nil {
			// Error from closing listeners, or context timeout:
			log.Printf("HTTP server Shutdown: %v", err)
		}
		log.Print("HTTP server Shutdown")
		close(idleConnsClosed)
		cancel()
	}()

	go SaveToFileTimer(ctx, mapMetric, cfg)

	log.Print("Запуск сервера HTTP")

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}

	<-idleConnsClosed

	//SaveMapMetric(mapMetric, cfg)
	log.Print("HTTP server close")

}

func SaveToFileTimer(ctx context.Context, m model.MapMetric, cfg *config.Config_Server) {
	if cfg.StoreInterval == 0 {
		return
	}
	for {

		select {
		case <-ctx.Done():
			return
		default:
			m.SaveToFile()
		}
		time.Sleep(time.Duration(cfg.StoreInterval) * time.Second)

	}
}
