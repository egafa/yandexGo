package handler

import (
	"fmt"
	"html/template"
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
