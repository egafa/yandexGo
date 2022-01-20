package model

import (
	"encoding/json"
	"log"
	"os"
)

type Metric interface {
	SaveGaugeVal(nameMetric string, value float64)
	GetGaugeVal(nameMetric string) (float64, bool)
	SaveCounterVal(nameMetric string, value int64)
	GetCounterVal(nameMetric string) (int64, bool)
	GetGaugetMetricTemplate() GaugeTemplateMetric
	GetCounterMetricTemplate() CounterTemplateMetric
}

type MapMetric struct {
	GaugeData   map[string]float64
	CounterData map[string]int64
	flagSave    bool
	fileName    string
}

type MapMetricToSave struct {
	GaugeData   map[string]float64
	CounterData map[string]int64
}

type GaugeTemplateMetric struct {
	Typemetric string
	Data       map[string]float64
}
type CounterTemplateMetric struct {
	Typemetric string
	Data       map[string]int64
}

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func NewMapMetric() MapMetric {
	mapMetricVal := MapMetric{}
	mapMetricVal.GaugeData = make(map[string]float64)
	mapMetricVal.CounterData = make(map[string]int64)
	return mapMetricVal
}

func (m MapMetric) SetFileName(fileName string) {
	m.flagSave = true
	m.fileName = fileName
}

func (m MapMetric) SaveToFile() error {
	var MapMetricToSave MapMetricToSave

	if !m.flagSave {
		return nil
	}

	file, err := os.Create(m.fileName)
	if err != nil {
		log.Println("Ошибка создания файла: ", m.fileName, err.Error())
		return err
	}
	defer file.Close()

	MapMetricToSave.CounterData = make(map[string]int64)
	for 
	copy(MapMetricToSave, m.CounterData)

	copy(target, source)

	encoder := json.NewEncoder(file)
	err = encoder.Encode(MapMetricToSave)
	if err != nil {
		log.Println("Ошибка сериализации: ", err.Error())
		return err
	}

	return nil
}

func (m MapMetric) SaveGaugeVal(nameMetric string, value float64) {
	m.GaugeData[nameMetric] = value
}

func (m MapMetric) GetGaugeVal(nameMetric string) (float64, bool) {
	res, ok := m.GaugeData[nameMetric]
	if ok {
		return res, true
	} else {
		return 0, false
	}

}

func (m MapMetric) SaveCounterVal(nameMetric string, value int64) {

	v, ok := m.CounterData[nameMetric]
	if ok {
		m.CounterData[nameMetric] = v + value
	} else {
		m.CounterData[nameMetric] = value
	}
}

func (m MapMetric) GetCounterVal(nameMetric string) (int64, bool) {

	v, ok := m.CounterData[nameMetric]
	if ok {
		return v, true
	} else {
		return 0, false
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

	res.Data = m.CounterData

	return res
}

func (m MapMetric) SetData(GaugeData map[string]float64, CounterData map[string]int64) {

	m.GaugeData = make(map[string]float64)
	m.CounterData = make(map[string]int64)

	for key, value := range GaugeData {
		m.GaugeData[key] = value
	}

	for key1, value1 := range CounterData {
		m.CounterData[key1] = value1
	}

}
