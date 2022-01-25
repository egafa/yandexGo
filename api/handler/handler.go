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

func bodyData(r *http.Request) (model.Metrics, error) {
	body, bodyErr := ioutil.ReadAll(r.Body)

	if bodyErr != nil {
		log.Print(" Ошибка открытия тела запроса " + bodyErr.Error())
		return model.Metrics{}, bodyErr
	}

	dataMetrics := model.Metrics{}
	jsonErr := json.Unmarshal(body, &dataMetrics)
	if jsonErr != nil {
		return model.Metrics{}, jsonErr
	}

	return dataMetrics, nil
}

func UpdateMetricHandlerChi(m model.Metric) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var dataMetrics model.Metrics
		var errConv error
		if r.Header.Get("Content-Type") == "application/json" {

			dataMetrics, errConv = bodyData(r)
			if errConv != nil {
				w.WriteHeader(http.StatusNotImplemented)
				http.Error(w, "Ошибка дессериализации", http.StatusBadRequest)
				return
			}
		} else {

			dataMetrics.ID = chi.URLParam(r, "typeMetric")
			dataMetrics.MType = chi.URLParam(r, "nammeMetric")
			valueMetric := chi.URLParam(r, "valueMetric")

			switch strings.ToLower(dataMetrics.MType) {
			case "gauge":
				f, err := strconv.ParseFloat(valueMetric, 64)
				if err == nil {
					dataMetrics.Value = &f
				}
				errConv = err

			case "counter":
				i, err := strconv.ParseInt(valueMetric, 10, 64)
				if err == nil {
					dataMetrics.Delta = &i
				}
				errConv = err
			default:
				w.WriteHeader(http.StatusBadRequest)
				http.Error(w, "Не определен тип метрики", http.StatusBadRequest)
				return
			}

		}

		if errConv != nil {
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, "Не определен тип метрики", http.StatusBadRequest)
		}

		switch strings.ToLower(dataMetrics.MType) {
		case "gauge":
			m.SaveGaugeVal(strings.ToLower(dataMetrics.ID), *dataMetrics.Value)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf("%v", *dataMetrics.Value)))

		case "counter":
			m.SaveCounterVal(strings.ToLower(dataMetrics.ID), *dataMetrics.Delta)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf("%v", *dataMetrics.Delta)))

		default:
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, "Не определен тип метрики", http.StatusBadRequest)
		}
	}
}

func ValueMetricHandlerChi(m model.Metric) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var logtext string

		if r.Method == http.MethodPost && r.Header.Get("Content-Type") == "application/json" {
			logtext = "*************** handler value Json " + r.URL.String()
			log.Println(logtext)

			dataMetrics, jsonErr := bodyData(r)
			if jsonErr != nil {
				w.WriteHeader(http.StatusNotImplemented)
				http.Error(w, "Ошибка дессериализации", http.StatusNotImplemented)
				log.Print(logtext + " Ошибка дессериализации")
				return
			}

			var ok bool
			switch strings.ToLower(dataMetrics.MType) {
			case "gauge":
				val, ok1 := m.GetGaugeVal(strings.ToLower(dataMetrics.ID))
				ok = ok1
				if ok {
					dataMetrics.Value = &val
				}
			case "counter":
				val, ok1 := m.GetCounterVal(strings.ToLower(dataMetrics.ID))
				ok = ok1
				if ok {
					dataMetrics.Delta = &val
				}

			default:
				w.WriteHeader(http.StatusNotFound)
				http.Error(w, "Не определен тип метрики", http.StatusNotFound)
				log.Print(logtext + " Не найдена метрика ")
				return
			}

			if ok {
				byt, err := json.Marshal(dataMetrics)
				if err == nil {
					w.Header().Set("Content-Type", "application/json")
					w.Write(byt)
					w.WriteHeader(http.StatusOK)
					log.Print(logtext + " Получено знаачене метрики" + string(byt))
					return
				}
			}

			w.WriteHeader(http.StatusNotFound)
			http.Error(w, "Не определен тип метрики", http.StatusNotFound)
			log.Print(logtext + " Не определен тип метрики")
			return
		}

		logtext = "*************** handler value plain " + r.URL.String()
		log.Println(logtext)

		typeMetric := chi.URLParam(r, "typeMetric")
		nameMetric := chi.URLParam(r, "nammeMetric")

		switch strings.ToLower(typeMetric) {
		case "gauge":
			val, ok := m.GetGaugeVal(strings.ToLower(nameMetric))
			if ok {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(fmt.Sprintf("%v", val)))
				return
			}
		case "counter":
			val, ok := m.GetCounterVal(strings.ToLower(nameMetric))
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
