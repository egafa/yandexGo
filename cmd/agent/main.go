package main

import (
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
	"syscall"
	"time"
)

func newRequest(urlReq string, method string, loger bool, infoLog *log.Logger) *http.Request {

	_, err := url.Parse(urlReq)
	if err != nil {
		log.Fatal(err)
	}

	req, _ := http.NewRequest(method, urlReq, nil)
	//req.Body.Close()

	if loger {
		infoLog.Printf("Request text: %s\n", urlReq)
	}

	return req
}

func formMetric(ctx context.Context, cfg cfg, namesMetric map[string]string, dataChannel chan *http.Request) {

	f, err := os.OpenFile("text.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644) // логи пока решил не переделывать, в 7 инкременте задание с логами связано, в памках икремента передалаю
	if err != nil {
		log.Println(err)
	}
	defer f.Close()

	infoLog := log.New(f, "INFO\t", log.Ldate|log.Ltime)

	urlUpdate := "http://%s/update/%s/%s/%v"

	for { //i := 0; i < 3; i++ {

		select {
		case <-ctx.Done():
			return
		default:
			{

				ms := runtime.MemStats{}
				runtime.ReadMemStats(&ms)

				v := reflect.ValueOf(ms)
				for key, typeNаme := range namesMetric {

					val := v.FieldByName(key).Interface()

					// сделал отдельную функцию для формирования запроса
					// не очень понял как прикрутить url.Parse для сборки url
					req := newRequest(fmt.Sprintf(urlUpdate, cfg.addrServer, typeNаme, key, val), http.MethodPost, cfg.log, infoLog)

					dataChannel <- req

				}

				dataChannel <- newRequest(fmt.Sprintf(urlUpdate, cfg.addrServer, "counter", "PollCount", 1), http.MethodPost, cfg.log, infoLog)

				dataChannel <- newRequest(fmt.Sprintf(urlUpdate, cfg.addrServer, "gauge", "RandomValue", rand.Float64()), http.MethodPost, cfg.log, infoLog)

				time.Sleep(time.Duration(cfg.intervalMetric) * time.Second)
			}
		}
	}
}

func sendMetric(ctx context.Context, dataChannel chan *http.Request, stopchanel chan int, cfg cfg) {
	var req *http.Request

	f, err := os.OpenFile("textreq.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644) // логи пока решил не переделывать, в 7 инкременте задание с логами связано, в памках икремента передалаю
	if err != nil {
		log.Println(err)
	}
	defer f.Close()

	infoLog := log.New(f, "INFO\t", log.Ldate|log.Ltime)

	client := &http.Client{}
	client.Timeout = time.Second * time.Duration(cfg.timeout)

	for { //i := 0; i < 40; i++ {

		select {
		case req = <-dataChannel:
			{

				resp, _ := client.Do(req)

				if cfg.log {
					infoLog.Printf("Request text: %s %v\n", req.URL, resp.Status)
				}
			}
		default:
			stopchanel <- 0
		}

	}

}

type cfg struct {
	addrServer     string
	log            bool
	intervalMetric int
	timeout        int
}

func main() {

	cfg := cfg{
		addrServer:     "127.0.0.1:8080",
		log:            false,
		intervalMetric: 4,
		timeout:        3,
	}

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
			namesMetric[strNаme] = "counter"
		case "float64":
			namesMetric[strNаme] = "gauge"
		default:
			continue
		}

	}

	ctx, cancel := context.WithCancel(context.Background())

	dataChannel := make(chan *http.Request, len(namesMetric)*100) //теперь в канале будет не url, а сам запрос. Мне кажется так лучше, когда в запросе будем заполнять Body
	stopchanel := make(chan int, 1)

	go formMetric(ctx, cfg, namesMetric, dataChannel)

	timer := time.NewTimer(2 * time.Second)
	<-timer.C

	go sendMetric(ctx, dataChannel, stopchanel, cfg)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	// Block until a signal is received.
	<-c

	cancel()

	<-stopchanel

}
