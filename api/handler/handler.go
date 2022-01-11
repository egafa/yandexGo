package handler

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/egafa/yandexGo/api/model"
	"github.com/go-chi/chi/v5"
)

func UpdateMetricHandlerChi(m model.Metric) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		typeMetric := chi.URLParam(r, "typeMetric")
		nameMetric := chi.URLParam(r, "nammeMetric")
		valueMetric := chi.URLParam(r, "valueMetric")

		var errConv error

		switch typeMetric {
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
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, "Не определена метрика", http.StatusBadRequest)
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

		typeMetric := chi.URLParam(r, "typeMetric")
		nameMetric := chi.URLParam(r, "nammeMetric")

		switch typeMetric {
		case "gauge":
			val, ok := m.GetGaugeVal(nameMetric)
			if ok {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(fmt.Sprintf("%v", val)))
			}
		case "counter":
			val, ok := m.GetCounterVal(nameMetric)
			if ok {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(fmt.Sprintf("%v", val)))
			}
		default:
			w.WriteHeader(http.StatusNotFound)
			http.Error(w, "Не найдена метрика", http.StatusNotFound)
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
