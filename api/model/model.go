package model

type Metric interface {
	SaveGaugeVal(nameMetric string, value float64)
	GetGaugeVal(nameMetric string) float64
	SaveCounterVal(nameMetric string, value int64)
	GetCounterVal(nameMetric string, num int64) int64
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

var MapMetricVal MapMetric

func GetMapMetricVal() MapMetric {
	return MapMetricVal
}

func InitMapMetricVal() {
	MapMetricVal = MapMetric{}
	MapMetricVal.GaugeData = make(map[string]float64)
	MapMetricVal.CounterData = make(map[string]int64)
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

func (m MapMetric) GetCounterVal(nameMetric string, num int64) (int64, bool) {

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
