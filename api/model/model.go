package model

import (
	"errors"
)

type Metric interface {
	SaveGaugeVal(nameMetric string, value float64)
	GetGaugeVal(nameMetric string) float64
	SaveCounterVal(nameMetric string, value int64)
	GetCounterVal(nameMetric string, num int64) int64
}

type MapMetric struct {
	GaugeData   map[string]float64
	CounterData map[string][]int64
}

type GaugeTemplateMetric struct {
	Typemetric string
	Data       map[string]float64
}
type CounterTemplateMetric struct {
	Typemetric string
	Data       map[string]int64
}

var MapMetricVal MapMetric

func GetMapMetricVal() MapMetric {
	return MapMetricVal
}

func InitMapMetricVal() {
	MapMetricVal = MapMetric{}
	MapMetricVal.GaugeData = make(map[string]float64)
	MapMetricVal.CounterData = make(map[string][]int64)
}

func (m MapMetric) SaveGaugeVal(nameMetric string, value float64) {
	m.GaugeData[nameMetric] = value
}

func (m MapMetric) GetGaugeVal(nameMetric string) (float64, error) {
	res, ok := m.GaugeData[nameMetric]
	if ok {
		return res, nil
	} else {
		return 0, errors.New("Не найдена метрика")
	}

}

func (m MapMetric) SaveCounterVal(nameMetric string, value int64) {
	var v []int64

	v, ok := m.CounterData[nameMetric]
	if ok {
		m.CounterData[nameMetric] = append(m.CounterData[nameMetric], value)
	}

	v = append(v, value)
	m.CounterData[nameMetric] = v
}

func (m MapMetric) GetCounterVal(nameMetric string, num int64) (int64, error) {
	var v []int64

	v, ok := m.CounterData[nameMetric]
	if ok {
		return v[len(v)-1], nil
	} else {
		return 0, errors.New("Не найдена метрика")
	}
}

func (m MapMetric) GetGaugetMetricTemplate() GaugeTemplateMetric {

	res := GaugeTemplateMetric{}

	res.Data = make(map[string]float64)

	res.Data = m.GaugeData
	res.Typemetric = "Gauge"

	return res
}

func (m MapMetric) GetCounterMetricTemplate() CounterTemplateMetric {

	res := CounterTemplateMetric{}

	res.Data = make(map[string]int64)
	res.Typemetric = "Counter"

	for name, v := range m.CounterData {
		res.Data[name] = v[len(v)-1]
	}

	return res
}
