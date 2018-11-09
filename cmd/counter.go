package cmd

import (
	"os"
	"path/filepath"

	"github.com/choria-io/prometheus-file-exporter/metrics"
)

func counter() error {
	metric := metrics.NewMetric(metricName, map[string]string{})
	_, err := os.Stat(filepath.Join(path, metric.FileName()))
	if err == nil {
		metric.Load(path)
	}

	metric.Incr(int(value))
	return metric.Save(path)
}
