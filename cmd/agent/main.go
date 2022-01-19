package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"syscall"
	"time"

	"encoding/json"

	"github.com/caarlos0/env/v6"

	"github.com/egafa/yandexGo/api/model"
)

type dataRequest struct {
	addr   string
	method string
	body   []byte
}

func newRequest(m interface{}, addr, method string) (dataRequest, error) {
	_, err := url.Parse(addr)
	if err != nil {
		log.Fatal("Ошибка парсера URL", err.Error())
	}

	r := dataRequest{}

	byt, err := json.MarshalIndent(m, "", "")
	if err != nil {
		return r, err
	}

	r.addr = addr
	r.method = method
	r.body = byt

	return r, nil
}

func formMetric(ctx context.Context, cfg cfg, namesMetric map[string]string, keysMetric []string, dataChannel chan []dataRequest) {

	urlUpdate := "http://%s/update/%s/%s/%v"

	for i := 0; i < 60; i++ {

		select {
		case <-ctx.Done():
			return
		default:
			{

				ms := runtime.MemStats{}
				runtime.ReadMemStats(&ms)

				sliceMetric := make([]dataRequest, len(keysMetric)+2)

				m := model.Metrics{}
				m.ID = "PollCount"
				m.MType = "counter"
				delta, _ := strconv.ParseInt("1", 10, 64)
				m.Delta = &delta

				req, err := newRequest(m, fmt.Sprintf(urlUpdate, cfg.addrServer, m.MType, m.ID, delta), http.MethodPost)
				if err == nil {
					sliceMetric[0] = req
				}

				m.ID = "RandomValue"
				m.MType = "gauge"
				delta, _ = strconv.ParseInt("0", 10, 64)
				m.Delta = &delta
				mValue := rand.Float64()
				m.Value = &mValue

				req, err = newRequest(m, fmt.Sprintf(urlUpdate, cfg.addrServer, m.MType, m.ID, mValue), http.MethodPost)
				if err == nil {
					sliceMetric[1] = req
				}

				v := reflect.ValueOf(ms)

				for i := 0; i < len(keysMetric); i++ {

					val := v.FieldByName(keysMetric[i]).Interface()
					typeNаme := namesMetric[keysMetric[i]]
					m := model.Metrics{}
					m.ID = keysMetric[i]
					m.MType = typeNаme

					var addr string
					if typeNаme == "gauge" {
						f, _ := strconv.ParseFloat(fmt.Sprintf("%v", val), 64)
						m.Value = &f
						addr = fmt.Sprintf(urlUpdate, cfg.addrServer, m.MType, m.ID, f)
					} else {
						i, _ := strconv.ParseInt(fmt.Sprintf("%v", val), 10, 64)
						m.Delta = &i
						addr = fmt.Sprintf(urlUpdate, cfg.addrServer, m.MType, m.ID, i)
					}

					req, err := newRequest(m, addr, http.MethodPost)
					if err == nil {
						sliceMetric[i+2] = req
					}

				}

				dataChannel <- sliceMetric
				//time.Sleep(time.Duration(cfg.pollInterval) * time.Second)
			}
		}
	}
}

func sendMetric(ctx context.Context, dataChannel chan []dataRequest, stopchanel chan int, cfg cfg) {
	var textReq []dataRequest

	client := &http.Client{}
	client.Timeout = time.Second * time.Duration(cfg.timeout)

	for { //i := 0; i < 40; i++ {

		select {
		case <-ctx.Done():
			return
		case textReq = <-dataChannel:
			{

				for j := 0; j < len(textReq); j++ {

					req, errReq := http.NewRequest(textReq[j].method, textReq[j].addr, bytes.NewBuffer(textReq[j].body))
					if errReq != nil {
						log.Fatal("Не удалось сформировать запрос ", errReq)
					}
					req.Header.Set("Content-Type", "application/json")

					_, err := client.Do(req)
					if err == nil {
						log.Println("Отправка запроса агента ", req.Method, " "+req.URL.String(), string(textReq[j].body))
					}
				}

				time.Sleep(time.Duration(cfg.reportInterval) * time.Second)
			}
		default:
			//stopchanel <- 0
			continue
		}

	}

}

type cfg struct {
	addrServer     string `env:"ADDRESS"`
	pollInterval   int    //`env:"POLL_INTERVAL"`
	reportInterval int    //`env:"REPORT_INTERVAL"`
	timeout        int
	dirname        string
}

func initconfig() cfg {
	var cfg cfg
	env.Parse(&cfg)

	if cfg.addrServer == "" {

		cfg.addrServer = "127.0.0.1:8080"
	}

	if cfg.pollInterval == 0 {
		cfg.pollInterval = 2
	}
	if cfg.reportInterval == 0 {
		cfg.reportInterval = 10
	}
	if cfg.timeout == 0 {
		cfg.timeout = 3
	}

	ex, err := os.Executable()
	if err != nil {
		cfg.dirname = ""
	} else {
		exPath := filepath.Dir(ex)
		cfg.dirname = exPath
	}

	return cfg
}

func main() {
	log.Println("Запуск агента")

	cfg := initconfig()

	namesMetric, keysMetric := namesMetric()
	log.Println("Массив метрик ", keysMetric)

	ctx, cancel := context.WithCancel(context.Background())

	dataChannel := make(chan []dataRequest, len(namesMetric)*100)

	go formMetric(ctx, cfg, namesMetric, keysMetric, dataChannel)

	timer := time.NewTimer(1 * time.Second)
	<-timer.C

	stopchanel := make(chan int, 1)
	//log.Println("Перед отправкой")
	go sendMetric(ctx, dataChannel, stopchanel, cfg)

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	// Block until a signal is received.
	<-sigint

	cancel()
	log.Println("Стоп агента")
	//<-stopchanel

}

func namesMetric() (map[string]string, []string) {
	ms := runtime.MemStats{}
	runtime.ReadMemStats(&ms)

	v := reflect.ValueOf(ms)
	typeOfS := v.Type()

	namesMetric := make(map[string]string)

	for i := 0; i < v.NumField(); i++ {
		typeNаme := reflect.TypeOf(v.Field(i).Interface()).String()
		strNаme := typeOfS.Field(i).Name
		switch typeNаme {
		case "uint64":
			namesMetric[strNаme] = "gauge" //"counter"
		case "uint32":
			namesMetric[strNаme] = "gauge" //"counter"
		case "float64":
			namesMetric[strNаme] = "gauge"
		default:
			continue
		}

	}

	keys := make([]string, 0, len(namesMetric))
	for k := range namesMetric {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return namesMetric, keys
}
