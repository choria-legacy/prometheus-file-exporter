package metrics

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// BasicMetric is the metric without any prometheus stuff
type BasicMetric struct {
	Type      string            `json:"type"`
	Name      string            `json:"name"`
	Labels    map[string]string `json:"labels"`
	Timestamp int64             `json:"timestamp"`
	Value     float64           `json:"value"`
}

// Metric is a on disk metric
type Metric struct {
	BasicMetric
	gauge *prometheus.GaugeVec
}

// NewMetric creates a new metric
func NewMetric(name string, labels map[string]string) *Metric {
	metric := &Metric{
		BasicMetric: BasicMetric{
			Type:      "gauge",
			Name:      name,
			Labels:    labels,
			Timestamp: time.Now().UnixNano(),
		},
	}

	return metric
}

// FileName returns the file name to store this metric in
func (b *Metric) FileName() string {
	if len(b.Labels) == 0 {
		return fmt.Sprintf("%s.json", b.Name)
	}

	return fmt.Sprintf("%s-%s.json", b.Name, strings.Join(b.LabelNames(), "-"))
}

// Incr increments the gauge
func (b *Metric) Incr(c int) {
	b.Value += float64(c)
	b.Timestamp = time.Now().UnixNano()
	b.Prometheus().WithLabelValues(b.LabelValues()...).Set(b.Value)
}

// Decr decrements the gauge
func (b *Metric) Decr(c int) {
	b.Incr(0 - c)
}

// Set sets the gauge value
func (b *Metric) Set(f float64) {
	b.Value = f
	b.Timestamp = time.Now().UnixNano()
	b.Prometheus().WithLabelValues(b.LabelValues()...).Set(f)
}

// Prometheus gets the GaugeVec
func (b *Metric) Prometheus() *prometheus.GaugeVec {
	if b.gauge != nil {
		return b.gauge
	}

	b.gauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: b.Name,
		Help: "Generated metric from " + b.FileName(),
	}, b.LabelNames())

	b.gauge.WithLabelValues(b.LabelValues()...).Set(b.Value)

	prometheus.Register(b.gauge)

	return b.gauge
}

// Unregister removes the metric from Prometheus
func (b *Metric) Unregister() {
	if b.gauge != nil {
		prometheus.Unregister(b.gauge)
	}
}

// LabelNames returns a sorted list of the label names
func (b *Metric) LabelNames() []string {
	keys := make([]string, 0, len(b.Labels))

	for k := range b.Labels {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

// LabelValues returns label values in the same order as LabelNames
func (b *Metric) LabelValues() []string {
	vals := make([]string, 0, len(b.Labels))

	for _, k := range b.LabelNames() {
		vals = append(vals, b.Labels[k])
	}

	return vals
}

// UnmarshalJSON is a custom json decoder that set the prometheus gauge value
func (b *Metric) UnmarshalJSON(in []byte) error {
	err := json.Unmarshal(in, &b.BasicMetric)
	if err != nil {
		return err
	}

	ts := b.Timestamp
	b.Set(b.Value)
	b.Timestamp = ts

	return nil
}

// Load loads the metric from the path
func (b *Metric) Load(path string) error {
	c, err := ioutil.ReadFile(filepath.Join(path, b.FileName()))
	if err != nil {
		return err
	}

	err = json.Unmarshal(c, b)
	if err != nil {
		return err
	}

	return nil
}

// Save saves the metric to disk in the given path
func (b *Metric) Save(path string) error {
	j, err := json.Marshal(b)
	if err != nil {
		return err
	}

	tempfile, err := ioutil.TempFile(path, "metric")
	if err != nil {
		return err
	}
	defer os.Remove(tempfile.Name())
	defer tempfile.Close()

	_, err = tempfile.Write(j)
	if err != nil {
		return err
	}

	err = os.Chmod(tempfile.Name(), 0644)
	if err != nil {
		return err
	}

	tempfile.Close()

	return os.Rename(tempfile.Name(), filepath.Join(path, b.FileName()))
}
