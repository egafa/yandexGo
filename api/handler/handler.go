package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/egafa/yandexGo/api/model"
	"github.com/go-chi/chi/v5"
)

func UpdateMetricHandlerChi(m model.Metric) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		typeMetric := chi.URLParam(r, "typeMetric")
		nameMetric := chi.URLParam(r, "nammeMetric")
		valueMetric := chi.URLParam(r, "valueMetric")

		if r.Header.Get("Content-Type") == "application/json" {

			body, bodyErr := ioutil.ReadAll(r.Body)
			if bodyErr != nil {
				w.WriteHeader(http.StatusNotImplemented)
				http.Error(w, "Не определен тип метрики", http.StatusNotImplemented)
				return
			}

			dataMetrics := model.Metrics{}
			jsonErr := json.Unmarshal(body, &dataMetrics)
			if jsonErr != nil {
				w.WriteHeader(http.StatusNotImplemented)
				http.Error(w, "Не определен тип метрики", http.StatusNotImplemented)
				return
			}

			if strings.ToLower(dataMetrics.MType) != "gauge" && strings.ToLower(dataMetrics.MType) != "counter" {
				w.WriteHeader(http.StatusNotImplemented)
				http.Error(w, "Не определен тип метрики", http.StatusNotImplemented)
				return
			}

			if strings.ToLower(dataMetrics.MType) == "gauge" {
				m.SaveGaugeVal(dataMetrics.ID, *dataMetrics.Value)
			}
			if strings.ToLower(dataMetrics.MType) == "counter" {
				m.SaveCounterVal(dataMetrics.ID, *dataMetrics.Delta)
			}

			w.WriteHeader(http.StatusOK)

			return

		}

		if strings.ToLower(typeMetric) != "gauge" && strings.ToLower(typeMetric) != "counter" {
			w.WriteHeader(http.StatusNotImplemented)
			http.Error(w, "Не определен тип метрики", http.StatusNotImplemented)
			return
		}

		if strings.ToLower(typeMetric) == "gauge" {
			f, err := strconv.ParseFloat(valueMetric, 64)

			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				http.Error(w, "Не определена метрика", http.StatusBadRequest)
				return
			}

			m.SaveGaugeVal(nameMetric, f)
		}

		if strings.ToLower(typeMetric) == "counter" {
			i, err := strconv.ParseInt(valueMetric, 10, 64)

			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				http.Error(w, "Не определена метрика", http.StatusBadRequest)
				return
			}

			m.SaveCounterVal(nameMetric, i)
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("application-type", "text/plain")
		w.Write([]byte(valueMetric))
	}
}

func ValueMetricHandlerChi(m model.Metric) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Header.Get("Content-Type") == "application/json" {

			body, bodyErr := ioutil.ReadAll(r.Body)
			if bodyErr != nil {
				w.WriteHeader(http.StatusNotImplemented)
				http.Error(w, "Не определен тип метрики", http.StatusNotImplemented)
				return
			}

			dataMetrics := model.Metrics{}
			jsonErr := json.Unmarshal(body, &dataMetrics)
			if jsonErr != nil {
				w.WriteHeader(http.StatusNotImplemented)
				http.Error(w, "Не определен тип метрики", http.StatusNotImplemented)
				return
			}

			if strings.ToLower(dataMetrics.MType) != "gauge" && strings.ToLower(dataMetrics.MType) != "counter" {
				w.WriteHeader(http.StatusNotImplemented)
				http.Error(w, "Не определен тип метрики", http.StatusNotImplemented)
				return
			}

			if strings.ToLower(dataMetrics.MType) == "gauge" {
				val, ok := m.GetGaugeVal(dataMetrics.ID)
				if ok {

					dataMetrics.Value = &val
					byt, err := json.Marshal(m)
					if err != nil {
						w.WriteHeader(http.StatusNotImplemented)
						http.Error(w, "Не определен тип метрики", http.StatusNotImplemented)
						return
					}

					w.Header().Set("Content-Type", "application/json")
					w.Write(byt)
					w.WriteHeader(http.StatusOK)

				} else {
					w.WriteHeader(http.StatusNotFound)
					http.Error(w, "Не найдена метрика", http.StatusNotFound)
				}
			}
			if strings.ToLower(dataMetrics.MType) == "counter" {
				//m.SaveCounterVal(dataMetrics.ID, *dataMetrics.Delta)
			}

			w.WriteHeader(http.StatusOK)

			return

		}

		typeMetric := chi.URLParam(r, "typeMetric")
		nameMetric := chi.URLParam(r, "nammeMetric")

		if typeMetric == "gauge" {
			val, ok := m.GetGaugeVal(nameMetric)
			if ok {
				w.WriteHeader(http.StatusOK)
				//w.Write([]byte(fmt.Sprintf("nameMetric %s is: %v\n", nameMetric, val)))
				w.Write([]byte(fmt.Sprintf("%v", val)))

			} else {
				w.WriteHeader(http.StatusNotFound)
				http.Error(w, "Не найдена метрика", http.StatusNotFound)
			}

		}

		if typeMetric == "counter" {

			val, ok := m.GetCounterVal(nameMetric, -1)
			if ok {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(fmt.Sprintf("%v", val)))
				//w.Write([]byte(fmt.Sprintf("nameMetric %s is: %v\n", nameMetric, val)))
			} else {
				w.WriteHeader(http.StatusNotFound)
				http.Error(w, "Не найдена метрика", http.StatusNotFound)
			}

		}

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
