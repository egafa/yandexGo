package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
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

	"github.com/caarlos0/env"

	"github.com/egafa/yandexGo/api/model"
)

func newRequest(m interface{}, addr, method string, loger bool) (*http.Request, error) {
	_, err := url.Parse(addr)
	if err != nil {
		log.Println("Ошибка парсера", err.Error())
		log.Fatal(err)
	}

	byt, err := json.MarshalIndent(m, "", "")
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequest(method, addr, bytes.NewBuffer(byt))
	req.Header.Set("Content-Type", "application/json")
	req.Body.Close()

	log.Println("Формирование запроса агента " + req.Method + "  " + req.URL.String() + string(byt))

	//if loger {
	//	infoLog.Printf("Request text: %s\n", addr+string(byt))
	//}

	return req, nil
}

func formMetric(ctx context.Context, cfg cfg, namesMetric map[string]string, keysMetric []string, dataChannel chan *http.Request) {

	/*
		f, err := os.OpenFile(cfg.dirlog+"text.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Println(err)
		}
		defer f.Close()
		infoLog := log.New(f, "INFO\t", log.Ldate|log.Ltime)
	*/

	addrServer := cfg.addrServer

	for i := 0; i < 50; i++ {

		select {
		case <-ctx.Done():
			return
		default:
			{

				ms := runtime.MemStats{}
				runtime.ReadMemStats(&ms)

				m := model.Metrics{}
				m.ID = "PollCount"
				m.MType = "counter"
				delta, _ := strconv.ParseInt("1", 10, 64)
				m.Delta = &delta

				addr := addrServer + "/update/counter/PollCount/" + "1"
				//req, err := newRequest(m, addr, http.MethodPost, cfg.log, infoLog)
				req, err := newRequest(m, addr, http.MethodPost, cfg.log)
				if err == nil {
					dataChannel <- req
				}

				m.ID = "RandomValue"
				m.MType = "gauge"
				delta, _ = strconv.ParseInt("0", 10, 64)
				m.Delta = &delta
				mValue := rand.Float64()
				m.Value = &mValue

				addr = addrServer + "/update/gauge/RandomValue/" + fmt.Sprintf("%v", mValue)
				req, err = newRequest(m, addr, http.MethodPost, cfg.log)
				if err == nil {
					dataChannel <- req
				}

				v := reflect.ValueOf(ms)
				for i := 0; i < len(keysMetric); i++ {
					//	namesMetric1[keys[i]] = namesMetric[keys[i]]

					//for key, typeNаme := range namesMetric {

					val := v.FieldByName(keysMetric[i]).Interface()
					typeNаme := namesMetric[keysMetric[i]]
					m := model.Metrics{}
					m.ID = keysMetric[i]
					m.MType = typeNаme
					if typeNаme == "gauge" {
						f, _ := strconv.ParseFloat(fmt.Sprintf("%v", val), 64)
						m.Value = &f
					} else {
						i, _ := strconv.ParseInt(fmt.Sprintf("%v", val), 10, 64)
						m.Delta = &i
						//continue
					}

					addr := addrServer + "/update/" + typeNаme + "/" + keysMetric[i] + "/" + fmt.Sprintf("%v", val)
					//req, err := newRequest(m, addr, http.MethodPost, cfg.log, infoLog)
					req, err := newRequest(m, addr, http.MethodPost, cfg.log)
					if err == nil {
						dataChannel <- req
					}

				}

				time.Sleep(time.Duration(cfg.pollInterval) * time.Second)
			}
		}
	}
}

func sendMetric(ctx context.Context, dataChannel chan *http.Request, stopchanel chan int, cfg cfg) {
	var textReq *http.Request
	//var m model.Metrics

	/*
		f, err := os.OpenFile(cfg.dirname+"\\textreq.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Println(err)
		}
		defer f.Close()

		infoLog := log.New(f, "INFO\t", log.Ldate|log.Ltime)
	*/

	client := &http.Client{}
	client.Timeout = time.Second * time.Duration(cfg.timeout)
	//log.Println("перед циклом " + cfg.addrServer)
	for { //i := 0; i < 40; i++ {

		select {
		case textReq = <-dataChannel:
			{

				resp, err := client.Do(textReq)
				if err != nil {
					log.Println("Отправка запроса агента Ошибка " + textReq.Method + "  " + textReq.URL.String() + err.Error())
					dataChannel <- textReq
				} else {

					respBody, errResp := ioutil.ReadAll(resp.Body)
					if errResp != nil {
						log.Println("Ошиибка получения тела ответа " + errResp.Error())
					}
					log.Println("Отправка запроса агента " + textReq.Method + "  " + textReq.URL.String() + " Ответ " + string(respBody))

				}

				//if cfg.log {
				//	infoLog.Printf("Request text: %s\n", textReq.URL)
				//}

				if err != nil {
					continue
				}

				//if cfg.log {
				//	infoLog.Printf("Status: " + resp.Status)
				//}
			}
		default:
			stopchanel <- 0

		}
		//time.Sleep(time.Duration(cfg.reportInterval) * time.Second)

	}

}

type cfg struct {
	addrServer     string
	log            bool
	pollInterval   int
	reportInterval int
	timeout        int
	dirname        string
	dirlog         string
}

func initconfig() cfg {
	cfg := cfg{}
	env.Parse(&cfg)

	if cfg.addrServer == "" {
		cfg.addrServer = "http://127.0.0.1:8080"
	}

	if cfg.pollInterval == 0 {
		cfg.pollInterval = 2
	}
	if cfg.reportInterval == 0 {
		cfg.reportInterval = 2
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

	if cfg.dirlog == "" {
		cfg.dirlog = "D:\\gafa\\Go\\yandexGo\\"
	}

	cfg.log = false

	return cfg
}

func main() {

	cfg := initconfig()

	namesMetric, keysMetric := namesMetric()
	log.Println("Массив метрик ", keysMetric)

	ctx, cancel := context.WithCancel(context.Background())

	//dataChannel := make(chan string, len(namesMetric)*100)
	dataChannel := make(chan *http.Request, len(namesMetric)*100)

	go formMetric(ctx, cfg, namesMetric, keysMetric, dataChannel)

	timer := time.NewTimer(1 * time.Second) // Горутину по отправке метрик создаем с задержкой в две секунды
	<-timer.C

	stopchanel := make(chan int, 1)
	//log.Println("Перед отправкой")
	go sendMetric(ctx, dataChannel, stopchanel, cfg)

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	// Block until a signal is received.
	<-sigint

	//timer = time.NewTimer(60 * time.Second)
	//<-timer.C

	cancel()

	<-stopchanel

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

	//namesMetric1 := make(map[string]string)
	//for i := 0; i < len(keys); i++ {
	//	namesMetric1[keys[i]] = namesMetric[keys[i]]
	//}

	return namesMetric, keys
}
