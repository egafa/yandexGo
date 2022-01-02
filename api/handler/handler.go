package handler

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/egafa/ytest/api/model"
	"github.com/go-chi/chi/v5"
)

func MetricHandler(m model.Metric) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ss := string(r.URL.Path)

		//http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
		words := strings.Split(ss, "/")
		if len(words) < 5 {
			w.Write([]byte("Ошибка запроса"))
			w.Write([]byte(http.StatusText(400)))
			return
		}

		nameMetric := words[3]
		strVal := words[4]
		strType := words[2]
		if strType == "gauge" {
			f, err := strconv.ParseFloat(strVal, 64)
			if err != nil {
				m.SaveGaugeVal(nameMetric, 0)
			}
			m.SaveGaugeVal(nameMetric, f)
		}

		if strType == "counter" {
			i, err := strconv.ParseInt(strVal, 10, 64)
			if err != nil {
				m.SaveCounterVal(nameMetric, 0)
			}
			m.SaveCounterVal(nameMetric, i)
		}

		//for idx, word := range words {
		//	w.Write([]byte(fmt.Sprintf("Word %d is: %s\n", idx, word)))
		//}

	}
}

func UpdateMetricHandlerChi(w http.ResponseWriter, r *http.Request) {
	var m model.MapMetric

	typeMetric := chi.URLParam(r, "typeMetric")
	nameMetric := chi.URLParam(r, "nammeMetric")
	valueMetric := chi.URLParam(r, "valueMetric")

	m = model.GetMapMetricVal()

	if typeMetric == "gauge" {
		f, err := strconv.ParseFloat(valueMetric, 64)
		if err != nil {
			m.SaveGaugeVal(nameMetric, 0)
		}
		m.SaveGaugeVal(nameMetric, f)
	}

	if typeMetric == "counter" {
		i, err := strconv.ParseInt(valueMetric, 10, 64)
		if err != nil {
			m.SaveCounterVal(nameMetric, 0)
		}
		m.SaveCounterVal(nameMetric, i)
	}

}

func ValueMetricHandlerChi(w http.ResponseWriter, r *http.Request) {
	var m model.MapMetric
	var nilerror *error = nil

	typeMetric := chi.URLParam(r, "typeMetric")
	nameMetric := chi.URLParam(r, "nammeMetric")

	m = model.GetMapMetricVal()

	if typeMetric == "gauge" {
		val, err := m.GetGaugeVal(nameMetric)
		if &err != nilerror {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf("nameMetric %s is: %v\n", nameMetric, val)))
		} else {
			http.Error(w, err.Error(), http.StatusNotFound)
		}

	}

	if typeMetric == "counter" {

		val, err := m.GetCounterVal(nameMetric, -1)
		if &err != nilerror {
			w.Write([]byte(fmt.Sprintf("nameMetric %s is: %v\n", nameMetric, val)))
		} else {
			http.Error(w, err.Error(), http.StatusNotFound)
		}

	}

}

func ListMetricsChi(w http.ResponseWriter, r *http.Request) {
	m := model.GetMapMetricVal()
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

}
