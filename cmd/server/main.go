package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	handler "github.com/egafa/yandexGo/api/handler"
	model "github.com/egafa/yandexGo/api/model"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type cfg struct {
	ADDRESS        string
	STORE_INTERVAL int
	STORE_FILE     string
	RESTORE        bool
}

func initconfig() cfg {
	cfg := cfg{}

	cfg.ADDRESS = "127.0.0.1:8080"
	cfg.STORE_INTERVAL = 5
	cfg.RESTORE = true
	cfg.STORE_FILE = "\\tmp\\devops-metrics-db.json"

	cfg_ADDRESS := os.Getenv("ADDRESS")
	if cfg_ADDRESS != "" {
		cfg.ADDRESS = cfg_ADDRESS
	}

	cfg_STORE_INTERVAL := os.Getenv("STORE_INTERVAL")
	if cfg_STORE_INTERVAL != "" {
		i, err := strconv.ParseInt(cfg_STORE_INTERVAL, 10, 0)

		if err == nil {
			cfg.STORE_INTERVAL = int(i)
		}
	}

	cfg_RESTORE := os.Getenv("RESTORE")
	if cfg_RESTORE != "" {
		if cfg_RESTORE == "0" {
			cfg.RESTORE = false
		} else {
			cfg.RESTORE = true
		}
	}
	cfg.RESTORE = true
	/*
		cfg.ADDRESS = os.Getenv("ADDRESS")
		cfg.STORE_FILE = os.Getenv("STORE_FILE")
		cfg.STORE_INTERVAL = os.Getenv("STORE_INTERVAL")
	*/

	ex, err := os.Executable()
	if err != nil {
		cfg.STORE_FILE = ""
	} else {
		exPath := filepath.Dir(ex)
		//nameSlach := filepath.ToSlash(exPath)
		//nameSlach = strings.ReplaceAll(nameSlach, "//", "/")
		//nameSlach = filepath.FromSlash(nameSlach) + cfg.STORE_FILE
		cfg.STORE_FILE = exPath + cfg.STORE_FILE
	}

	//log.Printf("cfg.STORE_FILE: %v", cfg.STORE_FILE)
	//*/

	return cfg
}

func SaveMapMetric(m model.Metric, cfg cfg) {

	file, err := os.Create(cfg.STORE_FILE)
	if err != nil {
		log.Fatalf("Ошибка создания файла: %v", cfg.STORE_FILE)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(m)
	if err != nil {
		log.Fatalf("Ошибка сериализации: %v", err)
	}

}

func LoadMapMetric(fileName string) (model.MapMetric, error) {
	var v model.MapMetric

	file, err := os.OpenFile(fileName, os.O_RDONLY, 0777)
	if err != nil {
		log.Printf("Ошибка открытия файла: %v %v", fileName, err)
		return v, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&v)
	if err != nil {
		log.Printf("Ошибка десериализации: %v", err)
		return v, err

	}

	return v, nil
}

func main() {

	cfg := initconfig()

	mapMetric := model.NewMapMetric()

	/*
		if cfg.RESTORE {
			v, err := LoadMapMetric(cfg.STORE_FILE)
			if err != nil {
				model.InitMapMetricVal()
			} else {
				model.InitMapMetricValData(v.GaugeData, v.CounterData)
			}
		}
	*/

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
		Addr:    cfg.ADDRESS,
	}

	/*
		go func() {
			for { //i := 0; i < 25; i++ {
				time.Sleep(time.Duration(cfg.STORE_INTERVAL) * time.Second)
				SaveMapMetric(mapMetric, cfg)
			}
		}()
	*/

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

	log.Print("Запуск сервера HTTP")

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}

	<-idleConnsClosed

	//SaveMapMetric(mapMetric, cfg)
	log.Print("HTTP server close")

}
