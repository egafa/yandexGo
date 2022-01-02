package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"syscall"
	"time"
)

func formMetric(ctx context.Context, cfg cfg, namesMetric map[string]string, dataChannel chan string) {

	f, err := os.OpenFile("text.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()

	infoLog := log.New(f, "INFO\t", log.Ldate|log.Ltime)
	addrServer := cfg.addrServer

	pollCount := 0
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

					addr := addrServer + "/update/" + typeNаme + "/" + key + "/" + fmt.Sprintf("%v", val)
					if cfg.log {
						infoLog.Printf("Request text: %s\n", addr)
					}
					dataChannel <- addr

				}
				pollCount++
				addr := addrServer + "/update/counter/PollCount/" + fmt.Sprintf("%v", pollCount)
				if cfg.log {
					infoLog.Printf("Request text: %s\n", addr)
				}
				dataChannel <- addr

				addr1 := addrServer + "/update/gauge/RandomValue/" + fmt.Sprintf("%v", rand.Float64())
				if cfg.log {
					infoLog.Printf("Request text: %s\n", addr)
				}
				dataChannel <- addr1

				time.Sleep(time.Duration(cfg.intervalMetric) * time.Second)
			}
		}
	}
}

func sendMetric(ctx context.Context, dataChannel chan string, stopchanel chan int, cfg cfg) {
	var textReq string
	f, err := os.OpenFile("textreq.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()

	infoLog := log.New(f, "INFO\t", log.Ldate|log.Ltime)

	client := &http.Client{}
	client.Timeout = time.Second * time.Duration(cfg.timeout)

	for { //i := 0; i < 40; i++ {

		select {
		case textReq = <-dataChannel:
			{
				req, _ := http.NewRequest(http.MethodGet, textReq, nil)
				resp, err := client.Do(req)
				if cfg.log {
					infoLog.Printf("Request text: %s\n", req.URL)
				}

				if err != nil {
					continue
				}

				if cfg.log {
					infoLog.Printf("Status: " + resp.Status)
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
		addrServer:     "http://127.0.0.1:8080",
		log:            true,
		intervalMetric: 4,
		timeout:        3,
	}

	ms := runtime.MemStats{}
	runtime.ReadMemStats(&ms)

	v := reflect.ValueOf(ms)
	typeOfS := v.Type()

	namesMetric := make(map[string]string)

	for i := 0; i < v.NumField(); i++ {
		typeNаme := fmt.Sprintf("%s", reflect.TypeOf(v.Field(i).Interface()))
		strNаme := fmt.Sprintf("%s", typeOfS.Field(i).Name)
		switch typeNаme {
		case "uint64":
			namesMetric[strNаme] = "counter" //append(namesMetric[strNаme], "counter")
		case "float64":
			namesMetric[strNаme] = "gauge" //append(namesMetric[strNаme], "gauge")
		default:
			continue
		}

	}
	//fmt.Println(names1)

	ctx, cancel := context.WithCancel(context.Background())

	dataChannel := make(chan string, len(namesMetric)*100)
	stopchanel := make(chan int, 1)
	go formMetric(ctx, cfg, namesMetric, dataChannel)

	timer := time.NewTimer(4 * time.Second) // создаём таймер
	<-timer.C

	go sendMetric(ctx, dataChannel, stopchanel, cfg)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	// Block until a signal is received.

	<-c

	cancel()

	<-stopchanel

}
