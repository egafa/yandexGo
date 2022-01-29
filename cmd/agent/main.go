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
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"syscall"
	"time"

	"encoding/json"

	"github.com/egafa/yandexGo/api/model"
	"github.com/egafa/yandexGo/config"
	"github.com/egafa/yandexGo/zipcompess"
)

type dataRequest struct {
	addr   string
	method string
	body   []byte
}

func newRequest(m interface{}, addr, method string, compress bool) (dataRequest, error) {
	_, err := url.Parse(addr)
	if err != nil {
		log.Fatal("Ошибка парсера URL", err.Error())
	}

	r := dataRequest{}

	byt, err := json.MarshalIndent(m, "", "")
	if err != nil {
		return r, err
	}

	if compress {
		byt, err = zipcompess.Compress(byt)
		if err != nil {
			log.Fatal("Ошибка сжатия данных", err.Error())
		}
	}

	r.addr = addr
	r.method = method
	r.body = byt

	return r, nil
}

func formMetric(ctx context.Context, cfg config.Config_Agent, namesMetric map[string]string, keysMetric []string, dataChannel chan []dataRequest) {

	urlUpdate := "http://%s/update/%s/%s/%v"

	for { //i := 0; i < 60; i++ { Здесь бесконечный цикл, это я тесты проводил

		select {
		case <-ctx.Done():
			return
		default:
			{

				ms := runtime.MemStats{}
				runtime.ReadMemStats(&ms)

				sliceMetric := make([]dataRequest, len(keysMetric)+2) //Оставил пока +2. Почему код хрупкий? Две наших метрики, поэтому + 2. Можно в массив keysMetric добавить наши метрики, но тогда при обходе этого массива  надо будет if вставлять

				m := model.Metrics{}
				m.ID = "PollCount"
				m.MType = "counter"
				delta := int64(1)
				m.Delta = &delta
				m.Hash = model.GetHash(m, cfg.Key)

				req, err := newRequest(m, fmt.Sprintf(urlUpdate, cfg.AddrServer, m.MType, m.ID, delta), http.MethodPost, cfg.Compress)
				if err == nil {
					sliceMetric[0] = req
				}

				m.ID = "RandomValue"
				m.MType = "gauge"
				delta, _ = strconv.ParseInt("0", 10, 64)
				m.Delta = &delta
				mValue := rand.Float64()
				m.Value = &mValue
				m.Hash = model.GetHash(m, cfg.Key)

				req, err = newRequest(m, fmt.Sprintf(urlUpdate, cfg.AddrServer, m.MType, m.ID, mValue), http.MethodPost, cfg.Compress)
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
						addr = fmt.Sprintf(urlUpdate, cfg.AddrServer, m.MType, m.ID, f)
					} else {
						i, _ := strconv.ParseInt(fmt.Sprintf("%v", val), 10, 64)
						m.Delta = &i
						addr = fmt.Sprintf(urlUpdate, cfg.AddrServer, m.MType, m.ID, i)
					}
					m.Hash = model.GetHash(m, cfg.Key)

					req, err := newRequest(m, addr, http.MethodPost, cfg.Compress)
					if err == nil {
						sliceMetric[i+2] = req
						//log.Println("Добавление запроса ", req.addr)
					}

				}

				dataChannel <- sliceMetric
				time.Sleep(time.Duration(cfg.PollInterval) * time.Second)
			}
		}
	}
}

func sendMetric(ctx context.Context, dataChannel chan []dataRequest, stopchanel chan int, cfg config.Config_Agent) {
	var textReq []dataRequest

	client := &http.Client{}
	client.Timeout = time.Second * time.Duration(cfg.Timeout)

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
					if cfg.Compress {
						req.Header.Set("Content-Encoding", "gzip")
					}

					_, err := client.Do(req)
					if err == nil {
						log.Println("Отправка запроса агента ", req.Method, " "+req.URL.String(), string(textReq[j].body))
					}
				}

				time.Sleep(time.Duration(cfg.ReportInterval) * time.Second)
			}
		default:
			//stopchanel <- 0
			continue
		}

	}

}

func main() {
	cfg := config.LoadConfigAgent()
	log.Println("Запуск агента.")
	log.Println("Сервер ", cfg.AddrServer, "PollInterval", cfg.PollInterval, "ReportInterval", cfg.ReportInterval, " Key ", cfg.Key)

	namesMetric, keysMetric := namesMetric()
	log.Println("Массив метрик ", keysMetric)

	ctx, cancel := context.WithCancel(context.Background())

	dataChannel := make(chan []dataRequest) //, len(namesMetric))

	go formMetric(ctx, *cfg, namesMetric, keysMetric, dataChannel)

	//time.Sleep(1 * time.Second)

	stopchanel := make(chan int, 1)
	go sendMetric(ctx, dataChannel, stopchanel, *cfg)

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
