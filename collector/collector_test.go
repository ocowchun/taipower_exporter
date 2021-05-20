package collector

import (
	"testing"

	"github.com/bmizerany/assert"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"gopkg.in/h2non/gock.v1"
)

type labelMap map[string]string

type MetricResult struct {
	labels     labelMap
	value      float64
	metricType dto.MetricType
}

func readMetric(m prometheus.Metric) MetricResult {
	pb := &dto.Metric{}
	m.Write(pb)
	labels := make(labelMap, len(pb.Label))
	for _, v := range pb.Label {
		labels[v.GetName()] = v.GetValue()
	}
	if pb.Gauge != nil {
		return MetricResult{labels: labels, value: pb.GetGauge().GetValue(), metricType: dto.MetricType_GAUGE}
	}
	if pb.Counter != nil {
		return MetricResult{labels: labels, value: pb.GetCounter().GetValue(), metricType: dto.MetricType_COUNTER}
	}
	if pb.Untyped != nil {
		return MetricResult{labels: labels, value: pb.GetUntyped().GetValue(), metricType: dto.MetricType_UNTYPED}
	}
	panic("Unsupported metric type")
}

func TestCollect(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		defer gock.Off() // Flush pending mocks after test execution

		gock.New("https://www.taipower.com.tw").
			Get("/d006/loadGraph/loadGraph/data/genloadareaperc.csv").
			Reply(200).
			BodyString("2021-05-19 22:40,1079.9,1131.9,1052.3,973.3,1140.5,1125.1,3.8,46.1\r\n")

		c := New()
		ch := make(chan prometheus.Metric, 100)
		c.Collect(ch)

		expecteds := []MetricResult{
			{labels: labelMap{"area": "northern_taiwan"}, value: float64(1079.9), metricType: dto.MetricType_GAUGE},
			{labels: labelMap{"area": "northern_taiwan"}, value: float64(1131.9), metricType: dto.MetricType_GAUGE},
			{labels: labelMap{"area": "central_taiwan"}, value: float64(1052.3), metricType: dto.MetricType_GAUGE},
			{labels: labelMap{"area": "central_taiwan"}, value: float64(973.3), metricType: dto.MetricType_GAUGE},
			{labels: labelMap{"area": "southern_taiwn"}, value: float64(1140.5), metricType: dto.MetricType_GAUGE},
			{labels: labelMap{"area": "southern_taiwn"}, value: float64(1125.1), metricType: dto.MetricType_GAUGE},
			{labels: labelMap{"area": "eastern_taiwan"}, value: float64(3.8), metricType: dto.MetricType_GAUGE},
			{labels: labelMap{"area": "eastern_taiwan"}, value: float64(46.1), metricType: dto.MetricType_GAUGE},
		}
		for _, expected := range expecteds {
			actual := readMetric(<-ch)
			assert.Equal(t, expected.value, actual.value)
			assert.Equal(t, expected.labels, actual.labels)
			assert.Equal(t, expected.metricType, actual.metricType)
		}
	})
}
