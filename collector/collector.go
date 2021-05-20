package collector

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func newMetric(metricName string, docString string) *prometheus.Desc {
	return prometheus.NewDesc(metricName, docString, []string{"area"}, make(map[string]string))
}

// Collector collects Taipower metrics. It implements prometheus.Collector interface.
type Collector struct {
	httpClient *http.Client
	metrics    map[string]*prometheus.Desc
	upMetric   prometheus.Gauge
	mutex      sync.Mutex
}

// New creates an Collector.
func New() *Collector {
	timeout := time.Second * 5
	httpClient := &http.Client{
		Timeout: timeout,
	}

	return &Collector{
		httpClient: httpClient,
		metrics: map[string]*prometheus.Desc{
			"power_consumption": newMetric("power_consumption", "Power Consumption"),
			"power_generation":  newMetric("power_generation", "Power Generation"),
		},
		upMetric: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "up",
			Help: "Status of the last metric scrape",
		}),
	}
}

const (
	dataEndpoint = "https://www.taipower.com.tw/d006/loadGraph/loadGraph/data/genloadareaperc.csv"
	serverUp     = 1
	serverDown   = 0
)

// Describe sends the super-set of all possible descriptors of Taipower metrics
// to the provided channel.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.upMetric.Desc()

	for _, m := range c.metrics {
		ch <- m
	}
}

// Collect fetches metrics from Taipower and sends them to the provided channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock() // To protect metrics from concurrent collects
	defer c.mutex.Unlock()

	data, err := c.scrape()
	if err != nil {
		c.upMetric.Set(serverDown)
		ch <- c.upMetric
		log.Printf("Error getting stats: %v", err)
		return
	}

	c.upMetric.Set(serverUp)
	ch <- c.upMetric

	if err := c.parseStats(ch, data); err != nil {
		log.Printf("Error parsing stats: %v", err)
	}
}

func (c *Collector) scrape() ([]byte, error) {
	resp, err := c.httpClient.Get(dataEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get %v: %v", dataEndpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expected %v response, got %v", http.StatusOK, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read the response body: %v", err)
	}
	return body, err
}

func (c *Collector) parseStats(ch chan<- prometheus.Metric, data []byte) error {
	dataStr := string(data)
	dataStr = strings.ReplaceAll(dataStr, "\r\n", "")

	parts := strings.Split(dataStr, ",")

	if len(parts) != 9 {
		return fmt.Errorf("invalid input %q", dataStr)
	}

	areas := []string{"northern_taiwan", "central_taiwan", "southern_taiwn", "eastern_taiwan"}
	for i, area := range areas {
		consumption, err := strconv.ParseFloat(parts[i*2+1], 64)
		if err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(c.metrics["power_consumption"],
			prometheus.GaugeValue, consumption, area)

		generation, err := strconv.ParseFloat(parts[i*2+2], 64)
		if err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(c.metrics["power_generation"],
			prometheus.GaugeValue, generation, area)
	}

	return nil
}
