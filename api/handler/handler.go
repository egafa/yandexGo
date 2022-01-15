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
	"github.com/go-chi/chi/v5"
)

func UpdateMetricHandlerChi(m model.Metric) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var logtext string

		if r.Header.Get("Content-Type") == "application/json" {
			logtext = "***************************** handler update Json " + r.URL.String()
			log.Println(logtext)

			body, bodyErr := ioutil.ReadAll(r.Body)
			defer r.Body.Close()
			if bodyErr != nil {
				w.WriteHeader(http.StatusNotImplemented)
				http.Error(w, "Ошибка открытия тела запроса", http.StatusBadRequest)
				log.Print(logtext + " Ошибка открытия тела запроса " + bodyErr.Error())
				return
			}

			dataMetrics := model.Metrics{}
			jsonErr := json.Unmarshal(body, &dataMetrics)
			if jsonErr != nil {
				w.WriteHeader(http.StatusNotImplemented)
				http.Error(w, "Ошибка дессериализации", http.StatusBadRequest)
				log.Print(logtext + " Ошибка дессериализации " + jsonErr.Error() + string(body))
				return
			}

			switch strings.ToLower(dataMetrics.MType) {
			case "gauge":
				m.SaveGaugeVal(dataMetrics.ID, *dataMetrics.Value)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(fmt.Sprintf("%v", *dataMetrics.Value)))
				log.Print(logtext + " Обработана метрика " + string(body))

			case "counter":
				m.SaveCounterVal(dataMetrics.ID, *dataMetrics.Delta)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(fmt.Sprintf("%v", *dataMetrics.Delta)))
				log.Print(logtext + " Обработана метрика " + string(body))

			default:
				w.WriteHeader(http.StatusBadRequest)
				http.Error(w, "Не определен тип метрики", http.StatusBadRequest)

				log.Print(logtext + " Не определен тип метрики " + string(body))
			}

			return

		}

		logtext = " ****************************** handler update plain " + r.URL.String()
		log.Println(logtext)

		typeMetric := chi.URLParam(r, "typeMetric")
		nameMetric := chi.URLParam(r, "nammeMetric")
		valueMetric := chi.URLParam(r, "valueMetric")

		var errConv error

		switch strings.ToLower(typeMetric) {
		case "gauge":
			f, err := strconv.ParseFloat(valueMetric, 64)
			if err == nil {
				m.SaveGaugeVal(nameMetric, f)
			}
			errConv = err

		case "counter":
			i, err := strconv.ParseInt(valueMetric, 10, 64)

			if err == nil {
				m.SaveCounterVal(nameMetric, i)
			}
			errConv = err

		default:
			w.WriteHeader(http.StatusNotImplemented)
			http.Error(w, "Не определен тип метрики", http.StatusNotImplemented)
			log.Print(logtext + " Не определен тип метрики ")
			return
		}

		if errConv != nil {
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, "Ошибка конвертации значения ", http.StatusBadRequest)
			log.Print(logtext + " Ошибка конвертации значения  ")
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("application-type", "text/plain")
		w.Write([]byte(valueMetric))
	}

}

func ValueMetricHandlerChi(m model.Metric) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var logtext string

		if r.Method == http.MethodPost && r.Header.Get("Content-Type") == "application/json" {
			logtext = "*************** handler value Json " + r.URL.String()
			log.Println(logtext)

			body, bodyErr := ioutil.ReadAll(r.Body)
			r.Body.Close()
			if bodyErr != nil {
				w.WriteHeader(http.StatusNotImplemented)
				http.Error(w, "Не определен тип метрики", http.StatusNotImplemented)
				log.Print(logtext + " Ошибка открытия тела запроса")
				return
			}

			dataMetrics := model.Metrics{}
			jsonErr := json.Unmarshal(body, &dataMetrics)
			if jsonErr != nil {
				w.WriteHeader(http.StatusNotImplemented)
				http.Error(w, "Ошибка дессериализации", http.StatusNotImplemented)
				log.Print(logtext + " Ошибка дессериализации" + string(body))
				return
			}

			if strings.ToLower(dataMetrics.MType) != "gauge" && strings.ToLower(dataMetrics.MType) != "counter" {
				w.WriteHeader(http.StatusNotImplemented)
				http.Error(w, "Не определен тип метрики", http.StatusNotImplemented)
				log.Print(logtext + " Не определен тип метрики")
				return
			}

			if strings.ToLower(dataMetrics.MType) == "gauge" {
				val, ok := m.GetGaugeVal(dataMetrics.ID)
				if ok {
					dataMetrics.Value = &val
				} else {
					w.WriteHeader(http.StatusNotFound)
					http.Error(w, "Не найдена метрика", http.StatusNotFound)
					log.Print(logtext + " Не найдена метрика " + string(body))
					return
				}
			}

			if strings.ToLower(dataMetrics.MType) == "counter" {
				val, ok := m.GetCounterVal(dataMetrics.ID)
				if ok {
					dataMetrics.Delta = &val
				} else {
					w.WriteHeader(http.StatusNotFound)
					http.Error(w, "Не найдена метрика", http.StatusNotFound)
					log.Print(logtext + " Не найдена метрика " + string(body))
					return
				}
			}

			var ok bool
			switch strings.ToLower(dataMetrics.MType) {
			case "gauge":
				val, ok := m.GetGaugeVal(dataMetrics.ID)
				if ok {
					dataMetrics.Value = &val
				}
			case "counter":
				val, ok := m.GetCounterVal(dataMetrics.ID)
				if ok {
					dataMetrics.Delta = &val
				}

			default:
				w.WriteHeader(http.StatusNotFound)
				http.Error(w, "Не определен тип метрики", http.StatusNotFound)
				log.Print(logtext + " Не найдена метрика " + string(body))
				return
			}

			if ok {
				byt, err := json.Marshal(dataMetrics)
				if err == nil {
					w.Header().Set("Content-Type", "application/json")
					w.Write(byt)
					w.WriteHeader(http.StatusOK)
					return
				}
			}

			w.WriteHeader(http.StatusNotFound)
			http.Error(w, "Не определен тип метрики", http.StatusNotFound)
			log.Print(logtext + " Не определен тип метрики" + string(body))
			return
		}

		logtext = "*************** handler value plain " + r.URL.String()
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

		w.WriteHeader(http.StatusNotFound)
		http.Error(w, "Не определен тип метрики", http.StatusNotFound)
		log.Print(logtext + " Не определен тип метрики " + typeMetric + "  " + nameMetric)

	}
}

func ListMetricsChiHandleFunc(m model.Metric) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		CounterData := m.GetCounterMetricTemplate()
		GaugeData := m.GetGaugetMetricTemplate()

		files := []string{
			"./internal/temptable.tmpl",
		}

		ts, err := template.ParseFiles(files...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

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
