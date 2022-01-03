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
	//var m model.MapMetric

	typeMetric := chi.URLParam(r, "typeMetric")
	nameMetric := chi.URLParam(r, "nammeMetric")
	valueMetric := chi.URLParam(r, "valueMetric")

	m := model.GetMapMetricVal()

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

func ValueMetricHandlerChi(w http.ResponseWriter, r *http.Request) {
	var m model.MapMetric

	typeMetric := chi.URLParam(r, "typeMetric")
	nameMetric := chi.URLParam(r, "nammeMetric")

	m = model.GetMapMetricVal()

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

	w.WriteHeader(http.StatusOK)
}
