package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/egafa/yandexGo/api/model"
	"github.com/egafa/yandexGo/config"
	"github.com/egafa/yandexGo/zipcompess"
	"github.com/go-chi/chi/v5"
)

func bodyData(r *http.Request) (model.Metrics, []byte, error) {
	var body []byte
	var bodyErr error

	defer r.Body.Close()
	if r.Header.Get("Content-Encoding") == "gzip" {
		body, bodyErr = zipcompess.Decompress(r.Body)
	} else {
		body, bodyErr = ioutil.ReadAll(r.Body)
	}

	if bodyErr != nil {
		log.Print(" Ошибка открытия тела запроса " + bodyErr.Error())
		return model.Metrics{}, nil, bodyErr
	}

	dataMetrics := model.Metrics{}
	jsonErr := json.Unmarshal(body, &dataMetrics)
	if jsonErr != nil {
		return model.Metrics{}, nil, jsonErr
	}

	return dataMetrics, body, nil
}

func UpdateMetricHandlerChi(m model.Metric, cfg *config.Config_Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var logtext string

		var dataMetrics model.Metrics
		var strBody string
		var jsonFlag bool
		var valueMetric string

		jsonFlag = false
		logtext = "******* handler update " + r.URL.Host + r.URL.String() + " Content-Encoding " + r.Header.Get("Content-Encoding")

		if r.Header.Get("Content-Type") == "application/json" {
			logtext = logtext + " ******* JSON"
			jsonFlag = true

			log.Println(logtext)

			dataMetrics1, body, jsonErr := bodyData(r)

			if jsonErr != nil {
				http.Error(w, "Ошибка дессериализации", http.StatusNotImplemented)
				log.Print(logtext + " Ошибка дессериализации " + jsonErr.Error() + string(body))
				return
			}

			h := model.GetHash(dataMetrics1, cfg.Key)
			if len(cfg.Key) > 0 && dataMetrics1.Hash != h {
				http.Error(w, "Хэш ключа не совпал", http.StatusNotImplemented)
				log.Print(logtext + " Хэш ключа не совпал " + string(body))
				return

			}

			dataMetrics = dataMetrics1
			strBody = string(body)

		} else {

			logtext = logtext + " ******* "
			log.Println(logtext)

			dataMetrics.ID = chi.URLParam(r, "nammeMetric")
			dataMetrics.MType = chi.URLParam(r, "typeMetric")
			valueMetric = chi.URLParam(r, "valueMetric")

		}

		var errConv error

		switch strings.ToLower(dataMetrics.MType) {
		case "gauge":
			if !jsonFlag {

				val, err := strconv.ParseFloat(valueMetric, 64)
				if err == nil {
					dataMetrics.Value = &val
				}
				errConv = err
			}

			if errConv == nil {
				errConv = m.SaveGaugeVal(dataMetrics.ID, *dataMetrics.Value)
			}

			if errConv == nil {
				log.Print(logtext + " Обработана метрика " + strBody)
				w.Write([]byte(fmt.Sprintf("%v", *dataMetrics.Value)))
			}

		case "counter":
			if !jsonFlag {
				val, err := strconv.ParseInt(valueMetric, 10, 64)
				if err == nil {
					dataMetrics.Delta = &val
				}
				errConv = err
			}
			if errConv == nil {
				m.SaveCounterVal(dataMetrics.ID, *dataMetrics.Delta)
				log.Print(logtext + " Обработана метрика " + strBody)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(fmt.Sprintf("%v", *dataMetrics.Delta)))
			}
		default:
			http.Error(w, "Не определен тип метрики", http.StatusNotImplemented)
			log.Print(logtext + " Не определен тип метрики ")
			return
		}

		if errConv != nil {
			http.Error(w, "Ошибка конвертации значения ", http.StatusBadRequest)
			log.Print(logtext + " Ошибка конвертации значения  ")
			return
		}

	}

}

func ValueMetricHandlerChi(m model.Metric, cfg *config.Config_Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var logtext string

		logtext = "******* Value " + r.URL.Host + r.URL.String() + " Content-Encoding " + r.Header.Get("Content-Encoding")

		if r.Method == http.MethodPost && r.Header.Get("Content-Type") == "application/json" {
			logtext = logtext + " ******* Json "
			log.Println(logtext)

			dataMetrics, body, jsonErr := bodyData(r)
			if jsonErr != nil {
				http.Error(w, "Ошибка дессериализации", http.StatusNotImplemented)
				log.Print(logtext + " Ошибка дессериализации" + string(body))
				return
			}

			var ok bool
			switch strings.ToLower(dataMetrics.MType) {
			case "gauge":
				val, ok1 := m.GetGaugeVal(dataMetrics.ID)
				ok = ok1
				if ok {
					dataMetrics.Value = &val
				}
			case "counter":
				val, ok1 := m.GetCounterVal(dataMetrics.ID)
				ok = ok1
				if ok {
					dataMetrics.Delta = &val
				}

			default:
				http.Error(w, "Не определен тип метрики", http.StatusNotFound)
				log.Print(logtext + " Не найдена метрика " + string(body))
				return
			}

			if ok {
				dataMetrics.Hash = model.GetHash(dataMetrics, cfg.Key)
				byt, err := json.Marshal(dataMetrics)
				if err == nil {
					w.Header().Set("Content-Type", "application/json")
					w.Write(byt)
					w.WriteHeader(http.StatusOK)
					log.Print(logtext + " Получено знаачене метрики" + string(byt))
					return
				}
			}

			http.Error(w, "Не определен тип метрики", http.StatusNotFound)
			log.Print(logtext + " Не определен тип метрики" + string(body))
			return
		}

		logtext = logtext + " ******* "
		log.Println(logtext)

		typeMetric := chi.URLParam(r, "typeMetric")
		nameMetric := chi.URLParam(r, "nammeMetric")

		switch strings.ToLower(typeMetric) {
		case "gauge":
			val, ok := m.GetGaugeVal(nameMetric)
			if ok {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(fmt.Sprintf("%v", val)))
				return
			}
		case "counter":
			val, ok := m.GetCounterVal(nameMetric)
			if ok {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(fmt.Sprintf("%v", val)))
				return
			}
		}

		http.Error(w, "Не определен тип метрики", http.StatusNotFound)
		log.Print(logtext + " Не определен тип метрики " + typeMetric + "  " + nameMetric)

	}
}

func ListMetricsChiHandleFunc(m model.Metric, cfg *config.Config_Server) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		logtext := "*******  List Metric " + r.URL.Host + r.URL.String() + " Content-Encoding " + r.Header.Get("Content-Encoding") + " Accept-Encoding " + r.Header.Get("Accept-Encoding")
		log.Println(logtext)

		CounterData := m.GetCounterMetricTemplate()
		GaugeData := m.GetGaugetMetricTemplate()

		files := []string{
			cfg.TemplateDir + "temptable.tmpl",
		}

		ts, err := template.ParseFiles(files...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			log.Println("Ошибка " + err.Error() + " парсинга шаблона " + cfg.TemplateDir + "temptable.tmpl")
			return
		}

		w.Header().Set("Content-Type", "text/html")

		err = ts.Execute(w, CounterData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		err = ts.Execute(w, GaugeData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func PingDBChiHandleFunc(m model.Metric) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		logtext := "*******  PingDB " + r.URL.Host + r.URL.String() + " Content-Encoding " + r.Header.Get("Content-Encoding") + " Accept-Encoding " + r.Header.Get("Accept-Encoding")
		log.Println(logtext)

		err := m.PingContext(r.Context())

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
