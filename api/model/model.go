package model

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
}

type GaugeTemplateMetric struct {
	Typemetric string
	Data       map[string]float64
}
type CounterTemplateMetric struct {
	Typemetric string
	Data       map[string]int64
}

func NewMapMetric() MapMetric {
	mapMetricVal := MapMetric{}
	mapMetricVal.GaugeData = make(map[string]float64)
	mapMetricVal.CounterData = make(map[string]int64)
	return mapMetricVal
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
