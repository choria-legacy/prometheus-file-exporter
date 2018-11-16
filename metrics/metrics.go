package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sync"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/rjeczalik/notify"
	"github.com/sirupsen/logrus"
)

var fileReadsGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "pfe_file_mtime_seconds",
	Help: "Unix mtime of read files",
}, []string{"file"})

var fileReadErrors = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "pfe_file_read_errors",
	Help: "Number of times files failed to read",
})

func init() {
	prometheus.MustRegister(fileReadErrors)
	prometheus.MustRegister(fileReadsGauge)
}

// Metrics is a collection of metrics found in a directory
type Metrics struct {
	sync.Mutex
	path    string
	metrics map[string]*Metric
	events  chan notify.EventInfo
	log     *logrus.Entry
}

// New creates a metric listener for a directory
func New(ctx context.Context, p string, watch bool, log *logrus.Entry) (*Metrics, error) {
	path, err := filepath.Abs(p)
	if err != nil {
		return nil, err
	}

	metrics := &Metrics{
		path:    path,
		metrics: make(map[string]*Metric),
		events:  make(chan notify.EventInfo, 1),
		log:     log.WithFields(logrus.Fields{"path": path}),
	}

	if watch {
		err = metrics.startNotify()
		if err != nil {
			return nil, fmt.Errorf("filesystem notifier setup failed: %s", err)
		}

		go metrics.watch(ctx)
	}

	metrics.initializeFromDisk()

	return metrics, nil
}

// Metrics retrieves the metric list
func (m *Metrics) Metrics() map[string]*Metric {
	return m.metrics
}

func (m *Metrics) loadMetric(f string) error {
	m.log.Debugf("Loading metric from %s", filepath.Base(f))

	m.Lock()
	defer m.Unlock()

	stat, err := os.Stat(f)
	if err != nil {
		return err
	}

	fileReadsGauge.WithLabelValues(f).Set(float64(stat.ModTime().Unix()))

	d, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}

	_, ok := m.metrics[f]
	if !ok {
		m.metrics[f] = &Metric{}
	}

	err = json.Unmarshal(d, m.metrics[f])
	if err != nil {
		return err
	}

	return nil
}

func (m *Metrics) removeMetric(f string) {
	m.Lock()
	defer m.Unlock()

	m.log.Debugf("Removing metric %s", filepath.Base(f))

	metric, ok := m.metrics[f]
	if !ok {
		return
	}

	metric.Unregister()
	delete(m.metrics, f)
}

func (m *Metrics) initializeFromDisk() {
	files, _ := filepath.Glob(filepath.Join(m.path, "*.json"))

	for _, file := range files {
		err := m.loadMetric(file)
		if err != nil {
			fileReadErrors.Inc()
			m.log.Errorf("Could not read metric %s: %s", file, err)
			continue
		}
	}

	return
}

func (m *Metrics) watch(ctx context.Context) {
	m.log.Infof("Watching %s for metric events", m.path)

	for {
		select {
		case event := <-m.events:
			m.log.Debugf("Handling event %#v", event)
			if m.shouldProcess(event.Path()) {
				_, err := os.Stat(event.Path())
				if err == nil {
					err := m.loadMetric(event.Path())
					if err != nil {
						fileReadErrors.Inc()
						m.log.Errorf("Could not handle metric update event for %s: %s", event.Path(), err)
					}
				} else {
					m.removeMetric(event.Path())
				}
			}

		case <-ctx.Done():
			notify.Stop(m.events)

			return
		}
	}
}

func (m *Metrics) shouldProcess(f string) bool {
	matched, err := regexp.MatchString("^[a-z].*.json$", filepath.Base(f))
	if err != nil {
		return false
	}

	return matched
}
