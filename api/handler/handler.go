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
			return
		}

		if errConv != nil {
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, "Не определена метрика", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("application-type", "text/plain")
		w.Write([]byte(valueMetric))
	}

}

func ValueMetricHandlerChi(m model.Metric) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		act := chi.URLParam(r, "act")
		if act != "value" {
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, "Не определена метрика", http.StatusBadRequest)
			return
		}

		typeMetric := chi.URLParam(r, "typeMetric")
		nameMetric := chi.URLParam(r, "nammeMetric")

		switch typeMetric {
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
		http.Error(w, "Не найдена метрика", http.StatusNotFound)

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
