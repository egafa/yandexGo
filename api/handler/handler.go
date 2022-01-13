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

		if r.Header.Get("Content-Type") == "application/json" {

			body, bodyErr := ioutil.ReadAll(r.Body)
			defer r.Body.Close()
			if bodyErr != nil {
				w.WriteHeader(http.StatusNotImplemented)
				http.Error(w, "Ошибка открытия тела запроса", http.StatusBadRequest)
				return
			}

			dataMetrics := model.Metrics{}
			jsonErr := json.Unmarshal(body, &dataMetrics)
			if jsonErr != nil {
				w.WriteHeader(http.StatusNotImplemented)
				http.Error(w, "Ошибка дессериализации", http.StatusBadRequest)
				return
			}

			switch strings.ToLower(dataMetrics.MType) {
			case "gauge":
				m.SaveGaugeVal(dataMetrics.ID, *dataMetrics.Value)

			case "counter":
				m.SaveCounterVal(dataMetrics.ID, *dataMetrics.Delta)

			default:
				w.WriteHeader(http.StatusBadRequest)
				http.Error(w, "Не определен тип метрики", http.StatusBadRequest)
			}

			return

		}

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

		if r.Method == http.MethodPost && r.Header.Get("Content-Type") == "application/json" {

			body, bodyErr := ioutil.ReadAll(r.Body)
			r.Body.Close()
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
				} else {
					w.WriteHeader(http.StatusNotFound)
					http.Error(w, "Не найдена метрика", http.StatusNotFound)
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
			return
		}

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
