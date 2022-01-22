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
	GaugeData   map[string]float64 `json:"GaugeData"`
	CounterData map[string]int64   `json:"CounterData"`
	FlagSave    bool               `json:"-"`
	FileName    string             `json:"-"`
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

func (m MapMetric) SetFileName(fname string) {
	m.FileName = fname
}

func (m *MapMetric) SetFlagSave(fl bool) {
	m.FlagSave = fl
}

func (m MapMetric) SaveToFile() error {
	//var MapMetricToSave MapMetric

	file, err := os.Create(m.FileName)
	if err != nil {
		log.Println("Ошибка создания файла: ", m.FileName, err.Error())
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(m)
	if err != nil {
		log.Println("Ошибка сериализации: ", err.Error())
		return err
	}

	return nil
}

func (m MapMetric) LoadFromFile() error {
	//var MapMetricToSave MapMetric

	if m.FileName == "" {
		return nil
	}

	file, err := os.OpenFile(m.FileName, os.O_RDONLY, 0777)
	if err != nil {
		log.Println("Ошибка открытия файла: ", m.FileName, err.Error())
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&m)
	if err != nil {
		log.Println("Ошибка десериализации: ", err.Error())
		return err

	}

	/*
		m = NewMapMetric()
		for k := range MapMetricToSave.CounterData {
			m.CounterData[k] = MapMetricToSave.CounterData[k]
		}

		for k := range MapMetricToSave.GaugeData {
			m.GaugeData[k] = MapMetricToSave.GaugeData[k]
		}
	*/

	return nil
}

func (m MapMetric) SaveGaugeVal(nameMetric string, value float64) {
	m.GaugeData[nameMetric] = value
	if m.FlagSave {
		m.SaveToFile()
	}
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

	if m.FlagSave {
		m.SaveToFile()
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
